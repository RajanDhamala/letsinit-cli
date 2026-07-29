import mongoose from "mongoose"

const connectDB = async () => {
     try {
          const mongoUrl = process.env.MONGODB_URL || "mongodb://127.0.0.1:27017/stackforge"
          await mongoose.connect(mongoUrl)
          console.log("connected to the mongodb server");
     } catch (Err) {
          console.log("failed to connect to the mongodb server", Err)
          process.exit(1)
     }
}


export default connectDB
