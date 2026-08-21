import { Outlet, createRootRoute } from "@tanstack/react-router"
import { QueryClientProvider } from "@tanstack/react-query"

import { queryClient } from "@/lib/query-client"

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  return (
    <QueryClientProvider client={queryClient}>
      <Outlet />
    </QueryClientProvider>
  )
}
