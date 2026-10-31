import { useState, useEffect } from "react"
import { me, logout, type AuthResponse } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { SignupPage } from "@/pages/SignupPage"
import { LoginPage } from "@/pages/LoginPage"

type View = "loading" | "login" | "signup" | "dashboard"

function App() {
  const [view, setView] = useState<View>("loading")
  const [user, setUser] = useState<AuthResponse | null>(null)

  useEffect(() => {
    me()
      .then((u) => {
        setUser(u)
        setView("dashboard")
      })
      .catch(() => setView("login"))
  }, [])

  function handleAuthSuccess(user: AuthResponse) {
    setUser(user)
    setView("dashboard")
  }

  async function handleLogout() {
    await logout()
    setUser(null)
    setView("login")
  }

  if (view === "loading") {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  if (view === "signup") {
    return <SignupPage onSuccess={handleAuthSuccess} onSwitchToLogin={() => setView("login")} />
  }

  if (view === "login") {
    return <LoginPage onSuccess={handleAuthSuccess} onSwitchToSignup={() => setView("signup")} />
  }

  return (
    <div className="mx-auto flex min-h-screen max-w-4xl flex-col px-4">
      <header className="flex items-center justify-between border-b py-4">
        <h1 className="text-xl font-semibold">SentinelPing</h1>
        <div className="flex items-center gap-4">
          <span className="text-muted-foreground text-sm">{user?.email}</span>
          <Button variant="outline" size="sm" onClick={handleLogout}>
            Log out
          </Button>
        </div>
      </header>

      <main className="flex flex-1 items-center justify-center">
        <p className="text-muted-foreground">
          No monitors yet. Phase 2 coming.
        </p>
      </main>
    </div>
  )
}

export default App
