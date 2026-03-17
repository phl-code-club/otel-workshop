import './instrumentation.js'
import express from "express";
import { connectToDB } from "./connection.js";
import authRouter from "./auth.js";
import { logger } from './logger.js';
import { PORT } from './const.js';

async function startServer() {
  await connectToDB()

  const app = express()
  app.use(express.json())
  app.use(authRouter)
  app.listen(PORT, () => logger.info(`server is running on port ${PORT}`, { port: PORT }))
}

startServer()
