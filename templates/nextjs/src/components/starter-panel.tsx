"use client"

import { useQuery } from "@tanstack/react-query"
import { CheckCircle2, Plus, RefreshCw, RotateCcw } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { api } from "@/lib/api"
import { useCounterStore } from "@/stores/use-counter-store"

interface StarterStatus {
  name: string
  ready: boolean
}

async function getStarterStatus() {
  const response = await api.get<StarterStatus>("/status")
  return response.data
}

export function StarterPanel() {
  const count = useCounterStore((state) => state.count)
  const increment = useCounterStore((state) => state.increment)
  const reset = useCounterStore((state) => state.reset)
  const statusQuery = useQuery({
    queryKey: ["starter-status"],
    queryFn: getStarterStatus,
  })

  const statusLabel = statusQuery.isPending
    ? "Checking the API..."
    : statusQuery.isError
      ? "The API request failed"
      : `${statusQuery.data.name} is ready`

  return (
    <Card className="mt-4 bg-card/80 backdrop-blur">
      <CardHeader>
        <CardTitle>State and data are connected</CardTitle>
        <CardDescription>
          This panel uses TanStack Query for the API request and Zustand for the counter.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-6 md:grid-cols-2">
        <div className="rounded-lg border bg-background/70 p-4">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium">TanStack Query</p>
              <p className="mt-1 text-sm text-muted-foreground">{statusLabel}</p>
            </div>
            {statusQuery.isSuccess && statusQuery.data.ready ? (
              <CheckCircle2 className="size-5 text-emerald-600" />
            ) : null}
          </div>
          <Button
            className="mt-4"
            disabled={statusQuery.isFetching}
            onClick={() => void statusQuery.refetch()}
            size="sm"
            type="button"
            variant="outline"
          >
            <RefreshCw className={statusQuery.isFetching ? "animate-spin" : ""} />
            Refresh status
          </Button>
        </div>

        <div className="rounded-lg border bg-background/70 p-4">
          <p className="text-sm font-medium">Zustand counter</p>
          <p className="mt-1 text-3xl font-semibold tabular-nums">{count}</p>
          <div className="mt-4 flex gap-2">
            <Button onClick={increment} size="sm" type="button">
              <Plus />
              Increment
            </Button>
            <Button onClick={reset} size="sm" type="button" variant="outline">
              <RotateCcw />
              Reset
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
