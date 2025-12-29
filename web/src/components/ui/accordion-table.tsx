import { useState, type ReactNode } from 'react'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

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
          <button
            onClick={allExpanded ? collapseAll : expandAll}
            className="text-xs text-muted-foreground hover:text-foreground"
          >
            {allExpanded ? collapseAllLabel : expandAllLabel}
          </button>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-muted-foreground">
              {hasChildren && <th className="w-8 py-2" />}
              {columns.map((col) => (
                <th
                  key={col.key}
                  className={cn('py-2 text-left font-medium', col.headerClassName)}
                >
                  {col.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {groups.map((group) => {
              const isExpanded = expanded.has(group.id)
              const hasChildRows = group.children.length > 0

              return (
                <>
                  <tr
                    key={group.id}
                    className={cn(
                      'border-b border-border/50 transition-colors',
                      hasChildRows && 'cursor-pointer hover:bg-muted/50'
                    )}
                    onClick={hasChildRows ? () => toggleGroup(group.id) : undefined}
                  >
                    {hasChildren && (
                      <td className="py-2 pr-1">
                        {hasChildRows && (
                          <span className="inline-flex items-center justify-center w-5 h-5">
                            {isExpanded ? (
                              <ChevronDown className="h-4 w-4 text-muted-foreground" />
                            ) : (
                              <ChevronRight className="h-4 w-4 text-muted-foreground" />
                            )}
                          </span>
                        )}
                      </td>
                    )}
                    {columns.map((col) => (
                      <td key={col.key} className={cn('py-2', col.className)}>
                        {col.render(group.summary)}
                      </td>
                    ))}
                  </tr>
                  {isExpanded &&
                    hasChildRows &&
                    childColumns &&
                    group.children.map((child, idx) => (
                      <tr
                        key={`${group.id}-child-${idx}`}
                        className="border-b border-border/30 bg-muted/20"
                      >
                        {hasChildren && <td className="py-1.5" />}
                        {childColumns.map((col) => (
                          <td key={col.key} className={cn('py-1.5 pl-4', col.className)}>
                            {col.render(child)}
                          </td>
                        ))}
                      </tr>
                    ))}
                </>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
