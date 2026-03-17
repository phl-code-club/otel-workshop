package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	l "phlcode.club/profile-service/logger"
)

var (
	tracer = otel.Tracer("profile-service")
	logger = l.GetLogger()
)

type contextKey string

var userKey = contextKey("user")

type UserResponse struct {
	User struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

type Profile struct {
	ID       int
	Username string
	Bio      string
	Location string
}

type Session struct {
	DB *sql.DB
}

func NewSession(db *sql.DB) *Session {
	return &Session{DB: db}
}

func getUserIDFromContext(ctx context.Context) (int, error) {
	val := ctx.Value(userKey)
	id, ok := val.(int)
	if !ok {
		return 0, errors.New("user id not a number")
	}

	return id, nil
}

func ValidateUser(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "validate-user")
		authHeader := r.Header.Get("authorization")
		if authHeader == "" {
			logger.WarnContext(ctx, "missing auth header")
			span.End()
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "http://auth-service:8000/user", nil)
		if err != nil {
			span.SetStatus(codes.Error, "error building auth request")
			span.RecordError(err)
			logger.ErrorContext(ctx, "error building auth request", "error", err)
			span.End()
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		req.Header.Set("authorization", authHeader)
		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
		res, err := client.Do(req)
		if err != nil {
			span.SetStatus(codes.Error, "error sending auth request")
			span.RecordError(err)
			logger.ErrorContext(ctx, "error sending auth request", "error", err)
			span.End()
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if res.StatusCode >= 400 {
			data, err := io.ReadAll(r.Body)
			body := string(data)
			if err != nil {
				// This is kinda ugly but I was sick of writing boilerplate lol
				body = err.Error()
			}
			span.SetStatus(codes.Error, "auth service returned an error")
			span.RecordError(errors.New(body))
			span.SetAttributes(attribute.String("error", string(body)))
			span.SetAttributes(attribute.Int("auth-request-status-code", res.StatusCode))
			span.SetAttributes(attribute.String("auth-request-status", res.Status))
			logger.ErrorContext(ctx, "auth service returned an error", "error", err)
			span.End()
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var responseBody UserResponse
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&responseBody)
		if err != nil {
			span.SetStatus(codes.Error, "unable to decode auth response")
			span.RecordError(err)
			span.SetAttributes(attribute.String("error", err.Error()))
			logger.ErrorContext(ctx, "unable to decode auth response", "error", err)
			span.End()
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		span.End()
		ctx = context.WithValue(r.Context(), userKey, responseBody.User.ID)
		r = r.WithContext(ctx)
		r.Context().Value(userKey)
		handler(w, r)
	}
}

func (s *Session) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "get-profile")
	defer span.End()
	id, err := getUserIDFromContext(ctx)
	if err != nil {
		span.SetStatus(codes.Error, "error getting user id from context")
		span.RecordError(err)
		logger.ErrorContext(ctx, "error getting user id from context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	row := s.DB.QueryRowContext(ctx, "SELECT id, username, bio, location FROM profiles WHERE id = $1", id)
	var profile Profile
	err = row.Scan(&profile.ID, &profile.Username, &profile.Bio, &profile.Location)
	if err != nil {
		span.SetStatus(codes.Error, "error querying profile from db")
		span.RecordError(err)
		logger.ErrorContext(ctx, "error querying profile from db", "error", err)
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error fetching profile", http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(profile)
	if err != nil {
		span.SetStatus(codes.Error, "error writing response")
		span.RecordError(err)
		logger.ErrorContext(ctx, "error writing response", "error", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}

func (s *Session) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "update-profile")
	defer span.End()
	id, err := getUserIDFromContext(r.Context())
	if err != nil {
		span.SetStatus(codes.Error, "error getting user id from context")
		span.RecordError(err)
		logger.ErrorContext(ctx, "error getting user id from context", "error", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var profileInput Profile
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&profileInput)
	if err != nil {
		span.SetStatus(codes.Error, "unable to read request body")
		span.RecordError(err)
		logger.ErrorContext(ctx, "unable to read request body", "error", err)
		http.Error(w, "Invalid profile input", http.StatusUnprocessableEntity)
		return
	}
	row := s.DB.QueryRowContext(ctx, `
	INSERT INTO profiles (id, username, bio, location) 
	VALUES ($1, $2, $3, $4) 
	ON CONFLICT (id) 
	DO UPDATE SET 
		username = EXCLUDED.username, 
		bio = EXCLUDED.bio, 
		location = EXCLUDED.location 
	RETURNING id, username, bio, location`,
		id,
		profileInput.Username,
		profileInput.Bio,
		profileInput.Location)

	var profile Profile
	err = row.Scan(&profile.ID, &profile.Username, &profile.Bio, &profile.Location)
	if err != nil {
		span.SetStatus(codes.Error, "unable to upsert profile")
		span.RecordError(err)
		logger.ErrorContext(ctx, "unable to upsert profile", "error", err)
		http.Error(w, "Error upserting profile", http.StatusInternalServerError)
		return
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(profile)
	if err != nil {
		span.SetStatus(codes.Error, "error writing response")
		span.RecordError(err)
		logger.ErrorContext(ctx, "error writing response", "error", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
}
