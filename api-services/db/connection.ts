import { Pool } from "pg";
import dotenv from "dotenv";

dotenv.config();

const pool = new Pool({
  host: process.env.DB_HOST,
  user: process.env.DB_USER,
  database: process.env.DB_NAME,
  password: process.env.DB_PASSWORD,
  port: parseInt(process.env.DB_PORT || "5432"),
});

export const connectToDB = async () => {
  try {
    await pool.connect();
    console.log("Database connected");
  } catch (err) {
    console.error("Database connection error:", err);
    process.exit(1);
  }
};

export default pool;
