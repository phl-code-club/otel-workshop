import db from "../db/connection";
import { Router, type Request, type Response } from "express";
import { z } from "zod";
import crypto from "crypto";
import { generateJWT } from "../utils/token";

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

    const passwordHash = crypto.createHash("sha256").update(safeParse.data.password).digest("hex")

    const result = await db.query(
      "INSERT INTO Users (email, password_hash) VALES ($1, $2) RETURNING *",
      [safeParse.data.email, passwordHash]
    )

    const user = result.rows[0]
    const token = generateJWT(user.id)

    return res.status(201).json({ user: user, token })
  }
  catch (err) {
    console.error("Error creating user:", err)
    return res.status(500).json({ error: "Internal server error" })
  }
}


const authRouter = Router()

authRouter.post("/signup", signUp)

export default authRouter;
