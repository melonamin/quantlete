import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Label } from '@/components/ui/label'
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
          <Label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={pullEnabled}
              onCheckedChange={(checked) => setPullEnabled(!!checked)}
              disabled={saving}
            />
            Enabled
          </Label>
        </div>

        <div className="flex items-center justify-between gap-4">
          <div>
            <p className="text-sm font-medium">Interval</p>
            <p className="text-xs text-muted-foreground">Uses the server's local time</p>
          </div>
          <Select
            value={pullSchedule}
            onValueChange={(value) => setPullSchedule(value as typeof pullSchedule)}
            disabled={saving || !pullEnabled}
          >
            <SelectTrigger className="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {scheduleOptions.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
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
          <Label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={pushEnabled}
              onCheckedChange={(checked) => setPushEnabled(!!checked)}
              disabled={saving}
            />
            Enabled
          </Label>
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
