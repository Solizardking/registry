import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:
          "bg-[#9945FF] text-white hover:bg-[#8a35f0] shadow-[0_0_16px_rgba(153,69,255,0.35)]",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline:
          "border border-[#9945FF]/35 bg-transparent text-zinc-100 hover:bg-[#9945FF]/10 hover:border-[#9945FF]/60 hover:text-white",
        secondary:
          "bg-[#14F195]/15 text-[#14F195] border border-[#14F195]/25 hover:bg-[#14F195]/25",
        ghost: "text-zinc-300 hover:bg-[#9945FF]/10 hover:text-white",
        link: "text-[#9945FF] underline-offset-4 hover:underline hover:text-[#b86bff]",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }

