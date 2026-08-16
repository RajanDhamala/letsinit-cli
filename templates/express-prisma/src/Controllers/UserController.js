import asyncHandler from "../Utils/AsyncHandler.js"
import ApiError from "../Utils/ApiError.js"
import ApiResponse from "../Utils/ApiResponse.js"
import prisma from "../Utils/PrismaProvider.js"
import { z } from "zod"
import { hashPassword,verifyPassword,CreateAccessToken,CreateRefreshToken } from "../Utils/Authutils.js"

const registerSchema = z.object({
  email: z.email(),
  fullname: z.string().trim().min(3).max(30),
  password: z.string().min(6)
}).strict();

const loginSchema = z.object({
  email: z.email(),
  password: z.string().min(6)
}).strict();

const RegisterUser = asyncHandler(async (req, res) => {
  const result = registerSchema.safeParse(req.body);

  if (!result.success) {
    throw new ApiError(400, "Invalid input", result.error.issues.map(issue => issue.message));
  }

  const value = result.data;

  const email = value.email.toLowerCase();

  const userExists = await prisma.user.findUnique({
    where: { email }
  });

  if (userExists) {
    throw new ApiError(400, "User with this email already exists");
  }

  const hashedPwd = await hashPassword(value.password);
  const newUser = await prisma.user.create({
    data: {
      email,
      fullname: value.fullname,
      password: hashedPwd
    },
    select: {
      id: true,
      email: true,
      fullname: true,
      createdAt: true
    }
  });

  return res.send(
    new ApiResponse(200, "User registered successfully", newUser)
  );
});

const LoginUser = asyncHandler(async (req, res) => {
  const result = loginSchema.safeParse(req.body);
  if (!result.success) {
    throw new ApiError(400, "Invalid credentials", result.error.issues.map(issue => issue.message));
  }

  const value = result.data;
  const email = value.email.toLowerCase();

  const existingUser = await prisma.user.findUnique({
    where: { email },
    select: {
      id: true,
      email: true,
      fullname: true,
      password: true
    }
  });

  if (!existingUser) {
    throw new ApiError(400, "Invalid credentials");
  }
  const isPasswordValid = await verifyPassword(value.password, existingUser.password);
  if (!isPasswordValid) {
    throw new ApiError(400, "Invalid credentials");
  }

  const accessToken = CreateAccessToken(
    existingUser.id,
    existingUser.email,
    existingUser.fullname
  );

  const refreshToken = CreateRefreshToken(
    existingUser.id,
    existingUser.email,
    existingUser.fullname
  );

  res.cookie("accessToken", accessToken, {
    httpOnly: true,
    secure: false,
    maxAge: 10 * 60 * 1000,
    path: "/"
  });

  res.cookie("refreshToken", refreshToken, {
    httpOnly: true,
    secure: false,
    maxAge: 7 * 24 * 60 * 60 * 1000,
    path: "/"
  });

  res.send(
    new ApiResponse(200, "User logged in successfully", {
      id: existingUser.id,
      fullname: existingUser.fullname,
      email: existingUser.email
    })
  );
});

const LogoutUser = asyncHandler(async (req, res) => {
  res.clearCookie("accessToken", {
    httpOnly: true,
    secure: false,
    path: "/"
  });

  res.clearCookie("refreshToken", {
    httpOnly: true,
    secure: false,
    path: "/"
  });

  res.send(new ApiResponse(200, "User logged out successfully"));
});

export {
    RegisterUser,LoginUser,LogoutUser
}
