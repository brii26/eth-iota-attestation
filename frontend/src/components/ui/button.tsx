import type { ComponentProps } from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva } from "class-variance-authority";
import type { VariantProps } from "class-variance-authority";

const buttonVariants = cva("inline-flex items-center justify-center whitespace-nowrap", {
  variants: { variant: { default: "", secondary: "secondary" } },
  defaultVariants: { variant: "default" },
});

type ButtonProps = ComponentProps<"button"> & VariantProps<typeof buttonVariants> & {
  asChild?: boolean;
};

export function Button({ className, variant, asChild = false, ...props }: ButtonProps) {
  const Component = asChild ? Slot : "button";
  return <Component className={buttonVariants({ variant, className })} data-slot="button" {...props} />;
}
