import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useTestNotification, useTestAllNotifications } from '@/lib/api/notifications'
import type {
  AppSettings,
  NotificationConfig,
  NotificationServiceConfig,
  NotificationEvents,
} from '@/lib/api/settings'
import { Loader2, Plus, Trash2, Send } from 'lucide-react'

const DEFAULT_NOTIFICATION_CONFIG: NotificationConfig = {
  enabled: false,
  services: [],
  events: {
    // Sync Events
    importComplete: true,

    // Achievements
    personalRecords: false,
    segmentPRs: false,
    eddingtonIncrease: false,
    powerRecords: false,
    goalComplete: false,

    // Training Load
    fatigueWarning: false,
    recoveryAlert: false,
    overtrainingRisk: false,

    // Summaries
    weeklyDigest: false,
    monthlyDigest: false,

    // Gear
    maintenanceDue: false,
    maintenanceSchedule: 'weekly',
    gearMilestones: false,
  },
}

type ServiceType = NotificationServiceConfig['type']

const SERVICE_TYPE_LABELS: Record<ServiceType, string> = {
  telegram: 'Telegram',
  smtp: 'Email (SMTP)',
  generic: 'Generic Webhook',
}

const SERVICE_CONFIG_FIELDS: Record<ServiceType, { key: string; label: string; type?: string }[]> =
  {
    telegram: [
      { key: 'bot_token', label: 'Bot Token', type: 'password' },
      { key: 'chat_id', label: 'Chat ID' },
    ],
    smtp: [
      { key: 'host', label: 'SMTP Host' },
      { key: 'port', label: 'Port' },
      { key: 'username', label: 'Username' },
      { key: 'password', label: 'Password', type: 'password' },
      { key: 'from', label: 'From Address' },
      { key: 'to', label: 'To Address' },
    ],
    generic: [
      { key: 'url', label: 'Webhook URL' },
      { key: 'method', label: 'Method' },
    ],
  }

interface NotificationsEditorProps {
  settings: AppSettings | undefined
  onSave: (s: AppSettings) => void
  saving: boolean
}

export function NotificationsEditor({ settings, onSave, saving }: NotificationsEditorProps) {
  const notifications = settings?.notifications ?? DEFAULT_NOTIFICATION_CONFIG

  const [enabled, setEnabled] = useState(notifications.enabled)
  const [services, setServices] = useState<NotificationServiceConfig[]>(notifications.services)
  const [events, setEvents] = useState<NotificationEvents>(notifications.events)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingService, setEditingService] = useState<NotificationServiceConfig | null>(null)

  const testNotification = useTestNotification()
  const testAllNotifications = useTestAllNotifications()

  const isDirty =
    enabled !== notifications.enabled ||
    JSON.stringify(services) !== JSON.stringify(notifications.services) ||
    JSON.stringify(events) !== JSON.stringify(notifications.events)

  const handleSave = () => {
    if (!settings) return
    onSave({
      ...settings,
      notifications: {
        enabled,
        services,
        events,
      },
    })
  }

  const handleAddService = () => {
    setEditingService({
      id: '',
      type: 'telegram',
      name: '',
      enabled: true,
      config: {},
    })
    setDialogOpen(true)
  }

  const handleEditService = (service: NotificationServiceConfig) => {
    setEditingService({ ...service })
    setDialogOpen(true)
  }

  const handleSaveService = () => {
    if (!editingService || !editingService.name.trim()) return

    if (editingService.id) {
      setServices(services.map((s) => (s.id === editingService.id ? editingService : s)))
    } else {
      const newService: NotificationServiceConfig = {
        ...editingService,
        id: crypto.randomUUID(),
      }
      setServices([...services, newService])
    }
    setDialogOpen(false)
    setEditingService(null)
  }

  const handleRemoveService = (id: string) => {
    setServices(services.filter((s) => s.id !== id))
  }

  const handleTestService = (serviceId: string) => {
    testNotification.mutate({ serviceId })
  }

  const handleTestAll = () => {
    testAllNotifications.mutate()
  }

  const updateEditingServiceConfig = (key: string, value: string) => {
    if (!editingService) return
    setEditingService({
      ...editingService,
      config: {
        ...editingService.config,
        [key]: value,
      },
    })
  }

  return (
    <div className="space-y-4">
      {/* Master Toggle */}
      <div className="rounded-md border border-border p-3">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="font-medium">Notifications</p>
            <p className="text-sm text-muted-foreground">
              Receive alerts for import completion and maintenance reminders
            </p>
          </div>
          <Label className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={enabled}
              onCheckedChange={(checked) => setEnabled(!!checked)}
              disabled={saving}
            />
            Enabled
          </Label>
        </div>
      </div>

      {/* Services Section */}
      {enabled && (
        <>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Services</p>
                <p className="text-sm text-muted-foreground">
                  Configure notification delivery channels
                </p>
              </div>
              <div className="flex gap-2">
                {services.filter((s) => s.enabled).length > 0 && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleTestAll}
                    disabled={testAllNotifications.isPending}
                  >
                    {testAllNotifications.isPending ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Send className="h-4 w-4" />
                    )}
                    <span className="ml-2">Test All</span>
                  </Button>
                )}
                <Button variant="outline" size="sm" onClick={handleAddService}>
                  <Plus className="h-4 w-4" />
                  <span className="ml-2">Add Service</span>
                </Button>
              </div>
            </div>

            {services.length === 0 ? (
              <div className="rounded-md border border-dashed border-border p-4 text-center text-sm text-muted-foreground">
                No notification services configured. Add one to start receiving notifications.
              </div>
            ) : (
              <div className="space-y-2">
                {services.map((service) => (
                  <div
                    key={service.id}
                    className="flex items-center justify-between gap-4 rounded-md border border-border p-3"
                  >
                    <div className="flex items-center gap-3 min-w-0">
                      <Checkbox
                        checked={service.enabled}
                        onCheckedChange={(checked) => {
                          setServices(
                            services.map((s) =>
                              s.id === service.id ? { ...s, enabled: !!checked } : s
                            )
                          )
                        }}
                        disabled={saving}
                      />
                      <div className="min-w-0">
                        <p className="font-medium truncate">{service.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {SERVICE_TYPE_LABELS[service.type]}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleTestService(service.id)}
                        disabled={!service.enabled || testNotification.isPending}
                        title="Send test notification"
                      >
                        {testNotification.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Send className="h-4 w-4" />
                        )}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleEditService(service)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRemoveService(service.id)}
                        className="text-destructive hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Events Section */}
          <div className="space-y-3">
            <div>
              <p className="font-medium">Events</p>
              <p className="text-sm text-muted-foreground">
                Choose which events trigger notifications
              </p>
            </div>

            {/* Sync Events */}
            <div className="rounded-md border border-border p-3">
              <p className="text-sm font-medium text-muted-foreground mb-3">Sync Events</p>
              <div className="space-y-2">
                <Label className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={events.importComplete}
                    onCheckedChange={(checked) =>
                      setEvents({ ...events, importComplete: !!checked })
                    }
                    disabled={saving}
                  />
                  Import Complete
                </Label>
                <p className="text-xs text-muted-foreground ml-6">
                  Notify when a data import finishes
                </p>
              </div>
            </div>

            {/* Achievements */}
            <div className="rounded-md border border-border p-3">
              <p className="text-sm font-medium text-muted-foreground mb-3">Achievements</p>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.personalRecords}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, personalRecords: !!checked })
                      }
                      disabled={saving}
                    />
                    Personal Records
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when setting a new best time on standard distances
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.segmentPRs}
                      onCheckedChange={(checked) => setEvents({ ...events, segmentPRs: !!checked })}
                      disabled={saving}
                    />
                    Segment PRs
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when setting a personal record on a segment
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.eddingtonIncrease}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, eddingtonIncrease: !!checked })
                      }
                      disabled={saving}
                    />
                    Eddington Increase
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when your Eddington number increases
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.powerRecords}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, powerRecords: !!checked })
                      }
                      disabled={saving}
                    />
                    Power Records
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when setting new peak power records
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.goalComplete}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, goalComplete: !!checked })
                      }
                      disabled={saving}
                    />
                    Goal Complete
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when completing weekly/monthly/yearly goals
                  </p>
                </div>
              </div>
            </div>

            {/* Training Load */}
            <div className="rounded-md border border-border p-3">
              <p className="text-sm font-medium text-muted-foreground mb-3">Training Load</p>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.fatigueWarning}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, fatigueWarning: !!checked })
                      }
                      disabled={saving}
                    />
                    Fatigue Warning
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Warn when training stress balance is very low
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.recoveryAlert}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, recoveryAlert: !!checked })
                      }
                      disabled={saving}
                    />
                    Recovery Alert
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when fully recovered from training
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.overtrainingRisk}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, overtrainingRisk: !!checked })
                      }
                      disabled={saving}
                    />
                    Overtraining Risk
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Warn when acute load exceeds chronic load significantly
                  </p>
                </div>
              </div>
            </div>

            {/* Summaries */}
            <div className="rounded-md border border-border p-3">
              <p className="text-sm font-medium text-muted-foreground mb-3">Summaries</p>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.weeklyDigest}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, weeklyDigest: !!checked })
                      }
                      disabled={saving}
                    />
                    Weekly Digest
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Send weekly activity summary on Mondays
                  </p>
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.monthlyDigest}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, monthlyDigest: !!checked })
                      }
                      disabled={saving}
                    />
                    Monthly Digest
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Send monthly activity summary on the 1st
                  </p>
                </div>
              </div>
            </div>

            {/* Gear */}
            <div className="rounded-md border border-border p-3">
              <p className="text-sm font-medium text-muted-foreground mb-3">Gear</p>
              <div className="space-y-4">
                <div className="space-y-3">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.maintenanceDue}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, maintenanceDue: !!checked })
                      }
                      disabled={saving}
                    />
                    Maintenance Due
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Remind about bike maintenance on a schedule
                  </p>

                  {events.maintenanceDue && (
                    <div className="flex items-center justify-between gap-4 ml-6">
                      <Label className="text-sm text-muted-foreground">Schedule</Label>
                      <Select
                        value={events.maintenanceSchedule}
                        onValueChange={(value: 'weekly' | 'monthly') =>
                          setEvents({ ...events, maintenanceSchedule: value })
                        }
                        disabled={saving}
                      >
                        <SelectTrigger className="w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="weekly">Weekly</SelectItem>
                          <SelectItem value="monthly">Monthly</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={events.gearMilestones}
                      onCheckedChange={(checked) =>
                        setEvents({ ...events, gearMilestones: !!checked })
                      }
                      disabled={saving}
                    />
                    Gear Milestones
                  </Label>
                  <p className="text-xs text-muted-foreground ml-6">
                    Notify when gear reaches 5,000 km milestones
                  </p>
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Save Button */}
      <div className="flex justify-end">
        <Button onClick={handleSave} disabled={saving || !isDirty}>
          Save
        </Button>
      </div>

      {/* Service Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingService?.id ? 'Edit Service' : 'Add Service'}</DialogTitle>
            <DialogDescription>Configure a notification delivery channel</DialogDescription>
          </DialogHeader>

          {editingService && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Name</Label>
                <Input
                  value={editingService.name}
                  onChange={(e) => setEditingService({ ...editingService, name: e.target.value })}
                  placeholder="My Telegram Bot"
                />
              </div>

              <div className="space-y-2">
                <Label>Type</Label>
                <Select
                  value={editingService.type}
                  onValueChange={(value: ServiceType) =>
                    setEditingService({
                      ...editingService,
                      type: value,
                      config: {},
                    })
                  }
                  disabled={!!editingService.id}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {(Object.keys(SERVICE_TYPE_LABELS) as ServiceType[]).map((type) => (
                      <SelectItem key={type} value={type}>
                        {SERVICE_TYPE_LABELS[type]}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {SERVICE_CONFIG_FIELDS[editingService.type].map((field) => (
                <div key={field.key} className="space-y-2">
                  <Label>{field.label}</Label>
                  <Input
                    type={field.type || 'text'}
                    value={editingService.config[field.key] || ''}
                    onChange={(e) => updateEditingServiceConfig(field.key, e.target.value)}
                    placeholder={field.label}
                  />
                </div>
              ))}
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSaveService} disabled={!editingService?.name.trim()}>
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
