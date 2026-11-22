import Image from "next/image"

export default function Home() {
  return (
    <div className="bg-noise font-sans grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20">
      <main className="flex flex-col gap-14 row-start-2 items-center">
        <div className="flex flex-col gap-10 items-center">
          <h1 className="font-display font-medium text-4xl text-center">
            Launchbox
          </h1>
          <Image
            className="dark:invert max-w-[144px]"
            src="/images/next.svg"
            alt="Next.js logo"
            width={180}
            height={38}
            priority
          />
          <div className="px-3 py-1 rounded-full backdrop-blur-sm bg-white/5 border border-foreground/15">
            <div className="flex items-center">
              <div
                data-text="Starter Kit"
                className="fx-text-glow relative font-mono font-medium text-sm"
              >
                <span className="fx-text-shimmer">Starter Kit</span>
              </div>
            </div>
          </div>
        </div>
        <ol className="font-mono list-inside list-decimal text-sm/6 text-center sm:text-left">
          <li className="mb-2 tracking-[-.01em]">
            Get started by editing{" "}
            <code className="bg-black/[.05] dark:bg-white/[.06] font-mono font-semibold px-1 py-0.5 rounded">
              app/(pages)/page.tsx
            </code>
            .
          </li>
          <li className="tracking-[-.01em]">
            Save and see your changes instantly.
          </li>
        </ol>
      </main>
    </div>
  )
}
