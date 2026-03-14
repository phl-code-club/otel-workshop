import db from "../db/connection";
import { Router, type Request, type Response } from "express";
import { z } from "zod";
import { generateJWT } from "../utils/token";
import { validateJWT } from "../middleware/validateToken";
import bcrypt from "bcrypt"

const userSchema = z.object({
  email: z.email({
    message: "A valid email is required"
  }),
  password: z.string().min(8, "Password must be at least 8 characters")
    .regex(/[A-Z]/, "Must contain at least one uppercase letter")
    .regex(/[0-9]/, "Must contain at least one number"),
})

async function signUp(req: Request, res: Response) {
  try {

    const safeParse = userSchema.safeParse(req.body)

    if (!safeParse.success) {
      return res.status(400).json({ error: safeParse.error.issues })
    }

    const passwordHash = await bcrypt.hash(safeParse.data.password, 10)

    if (typeof passwordHash != "string") {
      throw new Error("error hashing password")
    }

    const result = await db.query(
      "INSERT INTO Users (email, password_hash) VALES ($1, $2) RETURNING *",
      [safeParse.data.email, passwordHash]
    )

    const user = result.rows[0]

    const token = await generateJWT(user.id)
    if (!token) {
      throw new Error("error generating token")
    }

    return res.status(201).json({ user: user, token })
  }
  catch (err) {
    console.error("Error creating user:", err)
    return res.status(500).json({ error: "Internal server error" })
  }
}

async function signIn(req: Request, res: Response) {
  try {
    const safeParse = userSchema.safeParse(req.body)

    if (!safeParse.success) {
      return res.status(400).json({ error: safeParse.error.issues })
    }

    const result = await db.query(
      "SELECT * FROM Users WHERE email = $1",
      [safeParse.data.email]
    )

    if (result.rowCount === 0) {
      throw new Error("invalid email")
    }

    const user = result.rows[0]

    const match = await bcrypt.compare(safeParse.data.password, user.password_hash)
    if (!match) {
      throw new Error("invalid password")
    }

    const token = await generateJWT(user.id)
    if (!token) {
      throw new Error("error generating token")
    }

    return res.status(201).json({ user: user, token })
  }
  catch (err) {
    console.warn("Error signing in", err)
    return res.status(401).json({ error: "error authorizing user" })
  }
}


async function getUser(req: Request, res: Response) {
  try {
    const userID = res.locals.user
    if (!userID) {
      throw new Error("no userID")
    }

    const result = await db.query("SELECT * FROM Users WHERE id = $1",
      [userID])
    if (result.rowCount === 0) {
      throw new Error("User does not exist")
    }

    const user = result.rows[0]

    return res.status(201).json({ user: user })
  }
  catch (err) {
    console.warn("Error fetching user", err)
    return res.status(500).json({ error: "error authorizing user" })
  }
}

const authRouter = Router()

authRouter.post("/signup", signUp)
authRouter.post("/signin", signIn)
authRouter.get("/user", validateJWT, getUser)

export default authRouter;
