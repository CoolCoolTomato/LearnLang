# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Tauri v2 desktop application with a React 19 frontend. The project uses:
- **Frontend**: React 19, TypeScript, Vite, TailwindCSS v4, shadcn/ui components
- **Backend**: Rust with Tauri v2.10.3
- **Build Tool**: Vite 7 for frontend, Cargo for Rust backend

## Development Commands

### Frontend Development
```bash
npm run dev              # Start Vite dev server
npm run build            # Build frontend (TypeScript + Vite)
npm run lint             # Run ESLint
npm run format           # Format code with Prettier
npm run typecheck        # Run TypeScript type checking
npm run preview          # Preview production build
```

### Tauri Development
```bash
npm run tauri dev        # Start Tauri app in development mode
npm run tauri build      # Build production Tauri app
```

### Rust Backend
```bash
cd src-tauri
cargo build              # Build Rust backend
cargo test               # Run Rust tests
cargo clippy             # Run Rust linter
```

## Architecture

### Frontend Structure
- **src/**: React application source
  - **components/**: React components including shadcn/ui components in `ui/` subdirectory
  - **lib/**: Utility functions (e.g., `utils.ts` for cn() helper)
  - **App.tsx**: Main application component
  - **main.tsx**: React entry point
- **Path alias**: `@/` maps to `./src/`

### Backend Structure
- **src-tauri/**: Rust Tauri application
  - **src/lib.rs**: Main Tauri application logic with setup and plugin configuration
  - **src/main.rs**: Entry point that calls `app_lib::run()`
  - **Cargo.toml**: Rust dependencies and project configuration
  - **tauri.conf.json**: Tauri application configuration

### Styling
- TailwindCSS v4 with `@tailwindcss/vite` plugin
- shadcn/ui components using class-variance-authority and tailwind-merge
- Theme provider component for dark/light mode support

### Key Dependencies
- React 19 with React DOM
- Tauri plugins: tauri-plugin-log (for logging in debug mode)
- UI: lucide-react for icons, radix-ui primitives via shadcn
- Styling: TailwindCSS v4, clsx, tailwind-merge

## Important Notes

- The Rust backend uses a library crate pattern (`lib.rs` + `main.rs`)
- Logging is only enabled in debug builds via tauri-plugin-log
- TypeScript uses project references (tsconfig.app.json, tsconfig.node.json)
- ESLint configured for TypeScript with React hooks and React Refresh rules
