import { NextFunction, type Request, type Response } from "express"
import jwt from "jsonwebtoken"
import { SECRET } from "./const.js"
import { logger } from "./logger.js"
import { trace } from "@opentelemetry/api"

const tracer = trace.getTracer("validate-jwt-middleware");

export async function validateJWT(
  req: Request,
  res: Response,
  next: NextFunction
) {
  return tracer.startActiveSpan("validateJWT", (span) => {
    try {
      const authHeader = req.headers['authorization']
      if (!authHeader) {
        throw new Error("missing auth header")
      }

      const token = authHeader.split(" ")[1]
      if (!token) {
        throw new Error("missing token")
      }

      console.log(token)
      const decodedToken = jwt.verify(token, SECRET)
      if (typeof decodedToken === "string") {
        throw new Error("token should not be a string representation")
      }

      const userID = decodedToken.id

      res.locals.user = userID
      span.end()
      next()
    }
    catch (error) {
      span.setAttribute("error", error instanceof Error ? error.message : JSON.stringify(error))
      logger.warn("Error validating token", { error })
      span.end()
      return res.status(401).json({ error: "unauthorized" })
    }
  })
}
