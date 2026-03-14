import express from "express";
import * as dotenv from "dotenv";
import { validateJWT } from "../middleware/validateToken";
import authRouter from "../controllers/auth";

dotenv.config()

const app = express()
app.use(express.json())
app.use("/authservice", validateJWT, authRouter)

export default app
