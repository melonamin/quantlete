import type { ReactNode } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface WidgetWrapperProps {
  title: string
  children: ReactNode
  isLoading?: boolean
  action?: ReactNode
  className?: string
}

export function WidgetWrapper({
  title,
  children,
  isLoading = false,
  action,
  className = '',
}: WidgetWrapperProps) {
  return (
    <Card className={className}>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {action}
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <div className="space-y-2">
            <Skeleton className="h-8 w-24" />
            <Skeleton className="h-4 w-32" />
          </div>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  )
}
