import express from "express";
import cookieParser from "cookie-parser";
import dotenv from "dotenv";
import cors from "cors";
import UserRouter from "./src/Routes/UserRoute.js";

dotenv.config();
const app = express();

app.use(cors({
  origin: process.env.FRONTEND_URL || "http://localhost",
  credentials: true
}));
app.use(cookieParser());
app.use(express.urlencoded({ extended: true }));
app.use(express.json());

app.get("/", (req, res) => {
  res.send("Server is up and running");
});

app.get("/health", (req, res) => {
  res.status(200).json({ status: "ok" });
});

app.use("/users", UserRouter);

app.use((err, req, res, next) => {
  const statusCode = Number.isInteger(err.statusCode) ? err.statusCode : 500;

  res.status(statusCode).json({
    success: false,
    message: statusCode === 500 ? "Internal Server Error" : err.message,
    errors: Array.isArray(err.errors) ? err.errors : []
  });
});

export default app;
