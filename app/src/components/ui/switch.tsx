import * as React from "react"
import { cn } from "@/lib/utils"

function Switch({
  className,
  checked = false,
  disabled,
  onCheckedChange,
  ...props
}: Omit<React.ComponentProps<"button">, "onChange"> & {
  checked?: boolean
  onCheckedChange?: (checked: boolean) => void
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      data-state={checked ? "checked" : "unchecked"}
      className={cn(
        "relative inline-flex h-6 w-11 shrink-0 items-center rounded-full border border-transparent bg-input transition-colors outline-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary",
        className
      )}
      onClick={() => onCheckedChange?.(!checked)}
      {...props}
    >
      <span
        data-state={checked ? "checked" : "unchecked"}
        className="pointer-events-none block size-5 rounded-full bg-background shadow-sm transition-transform data-[state=checked]:translate-x-5 data-[state=unchecked]:translate-x-0"
      />
    </button>
  )
}

export { Switch }
