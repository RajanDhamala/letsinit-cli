import asyncHandler from "../Utils/AsyncHandler.js"
import ApiError from "../Utils/ApiError.js"
import ApiResponse from "../Utils/ApiResponse.js"
import User from "../Schemas/UserSchema.js"
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

  const normalizedEmail = value.email.toLowerCase();

  const userExists = await User.findOne({ email: normalizedEmail }).lean();
  if (userExists) {
    throw new ApiError(400, "User with this email already exists");
  }

  const pwd = await hashPassword(value.password);

  const newUser = new User({
    email: normalizedEmail,
    fullname: value.fullname,
    password: pwd,
  });

  await newUser.save();

  return res.send(
    new ApiResponse(200, "User registered successfully", {
      id: newUser._id,
      email: newUser.email,
      fullname: newUser.fullname,
    })
  );
});


const LoginUser = asyncHandler(async (req, res) => {
  const result = loginSchema.safeParse(req.body);
  if (!result.success) {
    throw new ApiError(400, "Invalid credentials", result.error.issues.map(issue => issue.message));
  }

  const value = result.data;
  const email = value.email.toLowerCase();

  const existingUser = await User.findOne({ email }).select("fullname _id email password");
  if (!existingUser) {
    throw new ApiError(400, "Invalid credentials");
  }
  const isValid = await verifyPassword(value.password, existingUser.password);
  if (!isValid) {
    throw new ApiError(400, "Invalid credentials");
  }

  const newAccessToken = CreateAccessToken(existingUser._id, existingUser.email, existingUser.fullname);
  const newRefreshToken = CreateRefreshToken(existingUser._id, existingUser.email, existingUser.fullname);

  res.cookie("accessToken", newAccessToken, {
    httpOnly: true,
    secure: false, 
    maxAge: 10 * 60 * 1000,
    path: "/",
  });

  res.cookie("refreshToken", newRefreshToken, {
    httpOnly: true,
    secure: false, 
    maxAge: 7 * 24 * 60 * 60 * 1000,
    path: "/",
  });

  res.send(
    new ApiResponse(200, "User logged in successfully", {
      id: existingUser._id,
      fullname: existingUser.fullname,
      email: existingUser.email
    })
  );
});


const LogoutUser=asyncHandler(async(req,res)=>{
    res.clearCookie("accessToken",{
    httpOnly: true,
    secure: false,
    path: "/",
    })
    res.clearCookie("refreshToken",{
    httpOnly: true,
    secure: false,
    path: "/",
    })
    res.send(new ApiResponse(200,'User logged out succesfully'))
})

export {
    RegisterUser,LoginUser,LogoutUser
}
