import * as React from "react"

interface LogoProps extends React.SVGProps<SVGSVGElement> {
  size?: number
}

export function Logo({ size = 24, className, ...props }: LogoProps) {
  return (
    <svg 
      width={size} 
      height={size} 
      viewBox="0 0 32 32" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      {...props}
    >
      <path 
        d="M9 7C9 5.34315 10.3431 4 12 4H20C21.6569 4 23 5.34315 23 7V11C23 12.6569 21.6569 14 20 14H16L12 17V14C10.3431 14 9 12.6569 9 11V7Z" 
        stroke="currentColor" 
        strokeWidth="2" 
        strokeLinecap="round" 
        strokeLinejoin="round" 
      />
      <path 
        d="M13 8H19M13 11H16" 
        stroke="currentColor" 
        strokeWidth="2" 
        strokeLinecap="round" 
      />
      
      <path 
        d="M16 18V28C16 28 12.5 26 6 26V16C12.5 16 16 18 16 18Z" 
        stroke="currentColor" 
        strokeWidth="2" 
        strokeLinecap="round" 
        strokeLinejoin="round" 
      />
      <path 
        d="M16 18V28C16 28 19.5 26 26 26V16C19.5 16 16 18 16 18Z" 
        stroke="currentColor" 
        strokeWidth="2" 
        strokeLinecap="round" 
        strokeLinejoin="round" 
      />
    </svg>
  )
}