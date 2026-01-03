// Re-export from generated types
import type { ActivityStreamResponse } from '@/lib/wasm/types.gen'
export type ActivityStream = ActivityStreamResponse

export { useActivities, useActivity, useActivityStreams } from '@/lib/data'
