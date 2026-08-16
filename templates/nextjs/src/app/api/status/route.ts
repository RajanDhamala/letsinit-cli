import { NextResponse } from "next/server"

export function GET() {
  return NextResponse.json({
    name: "Next.js starter",
    ready: true,
  })
}
