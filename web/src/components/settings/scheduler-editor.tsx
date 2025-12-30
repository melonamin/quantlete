import { useState } from 'react'
import { Button } from '@/components/ui/button'
import type { AppSettings } from '@/lib/api/settings'

interface SchedulerEditorProps {
  settings: AppSettings
  onSave: (s: AppSettings) => void
  saving: boolean
}

export function SchedulerEditor({ settings, onSave, saving }: SchedulerEditorProps) {
  const pull = settings.scheduler.pull
  const push = settings.scheduler.push

  const [pullEnabled, setPullEnabled] = useState(pull.enabled)
  const [pullSchedule, setPullSchedule] = useState(pull.schedule)
  const [pushEnabled, setPushEnabled] = useState(push.enabled)

  const dirty =
    pullEnabled !== pull.enabled || pullSchedule !== pull.schedule || pushEnabled !== push.enabled

  const scheduleOptions: { value: typeof pullSchedule; label: string }[] = [
    { value: 'midnight', label: 'Daily at midnight' },
    { value: 'every_6_hours', label: 'Every 6 hours' },
    { value: 'hourly', label: 'Every hour' },
  ]

  return (
    <div className="space-y-4">
      <div className="space-y-3 rounded-md border border-border p-3">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="font-medium">Pull</p>
            <p className="text-sm text-muted-foreground">
              Run an automatic incremental sync on a schedule
            </p>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={pullEnabled}
              onChange={(e) => setPullEnabled(e.target.checked)}
              disabled={saving}
            />
            Enabled
          </label>
        </div>

        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm font-medium">Interval</p>
            <p className="text-xs text-muted-foreground">Uses the server's local time</p>
          </div>
          <select
            className="rounded-md border border-border bg-background px-3 py-2 text-sm"
            value={pullSchedule}
            onChange={(e) => setPullSchedule(e.target.value as typeof pullSchedule)}
            disabled={saving || !pullEnabled}
          >
            {scheduleOptions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      <div className="rounded-md border border-border p-3">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="font-medium">Push (webhook)</p>
            <p className="text-sm text-muted-foreground">
              Process Strava webhook events in real time
            </p>
            <p className="text-xs text-muted-foreground">Endpoint: /api/v1/webhooks/strava</p>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={pushEnabled}
              onChange={(e) => setPushEnabled(e.target.checked)}
              disabled={saving}
            />
            Enabled
          </label>
        </div>
      </div>

      <div className="flex justify-end">
        <Button
          type="button"
          onClick={() => {
            onSave({
              ...settings,
              scheduler: {
                ...settings.scheduler,
                version: 2,
                pull: {
                  enabled: pullEnabled,
                  schedule: pullSchedule,
                },
                push: {
                  enabled: pushEnabled,
                },
              },
            })
          }}
          disabled={saving || !dirty}
        >
          Save
        </Button>
      </div>
    </div>
  )
}
