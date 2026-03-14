const jwt = require("jsonwebtoken")

export async function generateJWT(
  userId: number,
): Promise<string> {

  try {
    const key = process.env.SECRET
    if (!key) throw new Error("SECRET not defined")

    const payload = {
      iss: "auth-service",
      aud: "user-profile-service",
      id: userId
    }

    const token = await jwt.sign(payload, key, {
      algorithm: "HS256",
      expiresIn: "1 day"
    })

    return token
  }
  catch (err) {
    console.error("Error creating token", err)
    throw err
  }
}
