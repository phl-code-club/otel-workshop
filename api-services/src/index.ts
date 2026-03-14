
import app from "./server";
import * as dotenv from "dotenv";
import { connectToDB } from "../db/connection";

dotenv.config();

async function startServer() {
  await connectToDB()

  let port = process.env.PORT || 5000

  app.listen(port, () => console.log(`server is running on port ${port}`))
}

startServer()
