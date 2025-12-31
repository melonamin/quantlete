import {
  Hero,
  Features,
  Screenshots,
  Installation,
  OpenSource,
  Footer,
} from './components'

export default function App() {
  return (
    <div className="min-h-screen bg-background text-foreground noise-overlay scanline">
      <main>
        <Hero />
        <Features />
        <Screenshots />
        <Installation />
        <OpenSource />
      </main>
      <Footer />
    </div>
  )
}
