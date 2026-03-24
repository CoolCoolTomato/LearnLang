"use client"

import * as React from "react"
import { Check, ChevronsUpDown } from "lucide-react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface ModelComboboxProps {
  value: string
  onValueChange: (value: string) => void
  models: Array<{ id: string }>
  placeholder?: string
  disabled?: boolean
  className?: string
}

export function ModelCombobox({ value, onValueChange, models, placeholder, disabled = false, className }: ModelComboboxProps) {
  const [open, setOpen] = React.useState(false)
  const [inputValue, setInputValue] = React.useState(value)

  React.useEffect(() => {
    setInputValue(value)
  }, [value])

  const handleSelect = (currentValue: string) => {
    const newValue = currentValue === value ? "" : currentValue
    onValueChange(newValue)
    setInputValue(newValue)
    setOpen(false)
  }

  const handleInputChange = (newValue: string) => {
    setInputValue(newValue)
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && inputValue) {
      onValueChange(inputValue)
      setOpen(false)
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" role="combobox" aria-expanded={open} className={cn("justify-between", className)} disabled={disabled}>
          <span className="truncate">{value || placeholder || "Select"}</span>
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[--radix-popover-trigger-width] p-0">
        <Command onKeyDown={handleKeyDown}>
          <CommandInput placeholder="Search or type..." value={inputValue} onValueChange={handleInputChange} />
          <CommandList>
            <CommandEmpty>{inputValue ? `Press Enter to use "${inputValue}"` : "No results"}</CommandEmpty>
            <CommandGroup>
              {models.map((model) => (
                <CommandItem key={model.id} value={model.id} onSelect={handleSelect}>
                  <Check className={cn("mr-2 h-4 w-4", value === model.id ? "opacity-100" : "opacity-0")} />
                  {model.id}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
