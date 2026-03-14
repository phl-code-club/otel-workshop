import { NextFunction, type Request, type Response } from "express"
import jwt from "jsonwebtoken"

export async function validateJWT(
  req: Request,
  res: Response,
  next: NextFunction
) {

  try {
    const authHeader = req.headers['authorization']
    if (!authHeader) {
      throw new Error("missing auth header")
    }

    const token = authHeader.split(" ")[1]
    if (!token) {
      throw new Error("missing token")
    }

    const secret = process.env.SECRET
    if (!secret) {
      throw new Error("SECRET not defined")
    }

    const decodedToken = jwt.verify(token, secret)
    // req locals
    if (typeof decodedToken === "string") {
      /// TODO:what is this stringggg 
    }

    //TODO: store in locals

    next()
  }
  catch (err) {
    console.warn("Error validating token", err)
    return res.status(401).json({ error: "unauthorization" })
  }
}
