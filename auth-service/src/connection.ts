import { Client } from "pg";
import { logger } from "./logger.js";


const client = new Client({
  host: process.env.DB_HOST ?? "localhost",
  user: "phlcodeclub",
  database: "otel",
  password: "super-secret-password",
  port: 5432,
  ssl: false
});

export const connectToDB = async () => {
  try {
    await client.connect();
    logger.info("Database connected");
  } catch (error) {
    logger.error("Database connection error:", { error });
    process.exit(1);
  }
};

export default client;
