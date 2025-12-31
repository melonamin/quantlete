import { useState, type ReactNode } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from './button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './table'

export interface AccordionTableColumn<T> {
  key: string
  header: string
  render: (row: T) => ReactNode
  className?: string
  headerClassName?: string
}

export interface AccordionTableGroup<T, C> {
  id: string
  label: ReactNode
  summary: T
  children: C[]
}

interface AccordionTableProps<T, C> {
  columns: AccordionTableColumn<T>[]
  childColumns?: AccordionTableColumn<C>[]
  groups: AccordionTableGroup<T, C>[]
  defaultExpanded?: string[]
  expandAllLabel?: string
  collapseAllLabel?: string
  emptyMessage?: string
  className?: string
}

export function AccordionTable<T, C>({
  columns,
  childColumns,
  groups,
  defaultExpanded = [],
  expandAllLabel = 'Expand all',
  collapseAllLabel = 'Collapse all',
  emptyMessage = 'No data available.',
  className,
}: AccordionTableProps<T, C>) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set(defaultExpanded))

  const toggleGroup = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const expandAll = () => {
    setExpanded(new Set(groups.map((g) => g.id)))
  }

  const collapseAll = () => {
    setExpanded(new Set())
  }

  const allExpanded = expanded.size === groups.length && groups.length > 0
  const hasChildren = childColumns && groups.some((g) => g.children.length > 0)

  if (groups.length === 0) {
    return <div className="text-sm text-muted-foreground py-4">{emptyMessage}</div>
  }

  return (
    <div className={cn('space-y-2', className)}>
      {hasChildren && (
        <div className="flex justify-end">
          <Button
            variant="ghost"
            size="sm"
            onClick={allExpanded ? collapseAll : expandAll}
            className="text-xs text-muted-foreground h-auto py-1 px-2"
          >
            {allExpanded ? collapseAllLabel : expandAllLabel}
          </Button>
        </div>
      )}

      <Table>
        <TableHeader>
          <TableRow>
            {hasChildren && <TableHead className="w-8" />}
            {columns.map((col) => (
              <TableHead key={col.key} className={col.headerClassName}>
                {col.header}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {groups.map((group) => {
            const isExpanded = expanded.has(group.id)
            const hasChildRows = group.children.length > 0

            return (
              <>
                <TableRow
                  key={group.id}
                  className={cn(hasChildRows && 'cursor-pointer')}
                  onClick={hasChildRows ? () => toggleGroup(group.id) : undefined}
                >
                  {hasChildren && (
                    <TableCell className="pr-1">
                      {hasChildRows && (
                        <span className="inline-flex items-center justify-center w-5 h-5">
                          {isExpanded ? (
                            <ChevronDown className="h-4 w-4 text-muted-foreground" />
                          ) : (
                            <ChevronRight className="h-4 w-4 text-muted-foreground" />
                          )}
                        </span>
                      )}
                    </TableCell>
                  )}
                  {columns.map((col) => (
                    <TableCell key={col.key} className={col.className}>
                      {col.render(group.summary)}
                    </TableCell>
                  ))}
                </TableRow>
                {isExpanded &&
                  hasChildRows &&
                  childColumns &&
                  group.children.map((child, idx) => (
                    <TableRow
                      key={`${group.id}-child-${idx}`}
                      className="bg-muted/20"
                    >
                      {hasChildren && <TableCell />}
                      {childColumns.map((col) => (
                        <TableCell key={col.key} className={cn('pl-4', col.className)}>
                          {col.render(child)}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
              </>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
