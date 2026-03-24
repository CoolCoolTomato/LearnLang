import * as React from "react"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"

interface FieldProps {
  label: string
  htmlFor?: string
  description?: string
  children: React.ReactNode
  className?: string
}

export function Field({
  label,
  htmlFor,
  description,
  children,
  className,
}: FieldProps) {
  return (
    <div className={cn("grid gap-2.5", className)}>
      <div className="space-y-1">
        <Label htmlFor={htmlFor} className="text-sm font-medium">
          {label}
        </Label>
        {description ? (
          <p className="text-xs text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {children}
    </div>
  )
}