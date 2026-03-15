import jwt from "jsonwebtoken"
import { logger } from "./logger.js"
import { SECRET } from "./const.js"

export async function generateJWT(
  userId: number,
): Promise<string> {

  try {
    const payload = {
      iss: "auth-service",
      aud: "user-profile-service",
      id: userId
    }

    const token = jwt.sign(payload, SECRET, {
      algorithm: "HS256",
      expiresIn: "1 day"
    })

    return token
  }
  catch (error) {
    logger.error("Error creating token", { error })
    throw error
  }
}
