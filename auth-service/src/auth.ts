import db from "./connection.js";
import { Router, type Request, type Response } from "express";
import { z } from "zod";
import { generateJWT } from "./token.js";
import { validateJWT } from "./validateToken.js";
import bcrypt from "bcrypt"
import { logger } from "./logger.js";
import { SpanStatusCode, trace } from "@opentelemetry/api";

const tracer = trace.getTracer("auth-endpoint")

const userSchema = z.object({
  email: z.email({
    message: "A valid email is required"
  }),
  password: z.string().min(8, "Password must be at least 8 characters")
    .regex(/[A-Z]/, "Must contain at least one uppercase letter")
    .regex(/[0-9]/, "Must contain at least one number"),
})

async function signUp(req: Request, res: Response) {
  let span = tracer.startSpan("parse-body")
  try {

    const safeParse = userSchema.safeParse(req.body)

    if (!safeParse.success) {
      span.recordException(safeParse.error)
      span.setStatus({ code: SpanStatusCode.ERROR })
      logger.warn("User input error", { error: safeParse.error })
      span.end()
      return res.status(400).json({ error: safeParse.error.issues })
    }
    span.end()
    span = tracer.startSpan("hash-password")
    const passwordHash = await bcrypt.hash(safeParse.data.password, 10)

    if (typeof passwordHash != "string") {
      throw new Error("error hashing password")
    }
    span.end()
    const result = await db.query(
      "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, created_at;",
      [safeParse.data.email, passwordHash]
    )

    const user = result.rows[0]

    span = tracer.startSpan("generate-jwt")
    const token = await generateJWT(user.id)
    if (!token) {
      throw new Error("error generating token")
    }
    span.end()
    return res.status(201).json({ user: user, token })
  }
  catch (error) {
    let err: Error
    if (error instanceof Error) {
      err = error
    } else {
      err = new Error(`Unexpected error: ${JSON.stringify(error)}`, {
        cause: error
      })
    }
    span.recordException(err)
    span.setStatus({ code: SpanStatusCode.ERROR })
    logger.warn("Error creating user:", { error })
    span.end()
    return res.status(500).json({ error: "Internal server error" })
  }
}

async function signIn(req: Request, res: Response) {
  let span = tracer.startSpan("parse-body")
  try {
    const safeParse = userSchema.safeParse(req.body)

    if (!safeParse.success) {
      span.recordException(safeParse.error)
      span.setStatus({ code: SpanStatusCode.ERROR })
      logger.warn("User input error", { error: safeParse.error })
      span.end()
      return res.status(400).json({ error: safeParse.error.issues })
    }

    span.end()
    const result = await db.query(
      "SELECT id, email, created_at, password_hash FROM users WHERE email = $1",
      [safeParse.data.email]
    )

    if (result.rowCount === 0) {
      throw new Error("invalid email")
    }

    const user = result.rows[0]

    span = tracer.startSpan("compare-password")
    const match = await bcrypt.compare(safeParse.data.password, user.password_hash)
    if (!match) {
      throw new Error("invalid password")
    }

    span.end()
    span = tracer.startSpan("generate-jwt")
    const token = await generateJWT(user.id)
    if (!token) {
      throw new Error("error generating token")
    }
    delete user.password_hash
    span.end()
    return res.status(201).json({ user, token })
  }
  catch (error) {
    let err: Error
    if (error instanceof Error) {
      err = error
    } else {
      err = new Error(`Unexpected error: ${JSON.stringify(error)}`, {
        cause: error
      })
    }
    span.recordException(err)
    span.setStatus({ code: SpanStatusCode.ERROR })
    logger.warn("Error signing in", { error })
    span.end()
    return res.status(401).json({ error: "error authorizing user" })
  }
}


async function getUser(_req: Request, res: Response) {
  const span = trace.getActiveSpan();
  try {
    const userID = res.locals.user
    if (!userID) {
      throw new Error("no userID")
    }

    const result = await db.query("SELECT id, email, created_at FROM users WHERE id = $1",
      [userID])
    if (result.rowCount === 0) {
      throw new Error("User does not exist")
    }

    const user = result.rows[0]

    return res.status(201).json({ user: user })
  }
  catch (error) {
    let err: Error
    if (error instanceof Error) {
      err = error
    } else {
      err = new Error(`Unexpected error: ${JSON.stringify(error)}`, {
        cause: error
      })
    }
    if (span) {
      span.recordException(err)
      span.setStatus({ code: SpanStatusCode.ERROR })
    }
    console.warn("Error fetching user", { error })
    return res.status(401).json({ error: "error authorizing user" })
  }
}

const authRouter = Router()

authRouter.post("/signup", signUp)
authRouter.post("/signin", signIn)
authRouter.get("/user", validateJWT, getUser)

export default authRouter;
