const http = require("http")
const bodyParser = require("body-parser")
const compression = require("compression")
const cors = require("cors")
const express = require("express")
const next = require("next")

const REDIRECTS = require("./redirects.json")

const dev = process.env.NODE_ENV !== "production"
const port = parseInt(process.env.PORT || "3000", 10)
const app = next({ dev })
const handle = app.getRequestHandler()

async function startServer() {
  try {
    await app.prepare()

    const server = express()
    const httpServer = http.createServer(server)

    /* Global middlewares */
    server.use(cors())
    server.use(
      compression({
        threshold: 0,
        filter: () => true,
        level: 9,
      }),
    )
    server.use(bodyParser.json())
    server.use(bodyParser.urlencoded({ extended: true }))

    /* Remove trailing slash middleware */
    server.use((req, res, next) => {
      const hasQueryParamWithSlash = /\?[^]*\//.test(req.url)
      if (
        req.url.endsWith("/") &&
        req.url.length > 1 &&
        !hasQueryParamWithSlash
      ) {
        return res.redirect(301, req.url.slice(0, -1))
      }
      next()
    })

    /* Router for specific pages and redirects */
    const router = express.Router({
      caseSensitive: false,
      strict: false,
    })

    server.set("case sensitive routing", false)

    /* Bot Handler: reroute bots to /homeshare when the main page was too heavy to crawl. */
    // router.get("/", (req, res) => {
    //   const userAgent = req.headers["user-agent"] || ""

    //   const isBot =
    //     userAgent.startsWith("facebookexternalhit/1.1") ||
    //     userAgent === "Facebot" ||
    //     userAgent.startsWith("Twitterbot") ||
    //     userAgent.includes("TelegramBot")

    //   const targetPage = isBot ? "/homeshare" : "/"

    //   console.log({ userAgent, isBot, targetPage })

    //   return app.render(req, res, targetPage, req.query)
    // })

    /* Custom Pages */
    // router.get("/custompage", (req, res) => app.render(req, res, "/custompage", req.query))

    /* Redirects */
    REDIRECTS.forEach(({ path, redirectTo, status }) => {
      router.get(path, (req, res) => res.status(status).redirect(redirectTo))
    })

    /* Home post */
    router.post("/", (req, res) => app.render(req, res, "/", req.query))

    /* Only specific paths handled by router */
    server.use(router)

    /* 🔥 Catch all remaining requests here safely */
    server.use((req, res) => {
      return handle(req, res)
    })

    httpServer.listen(port, () => {
      console.log(
        `> Server running on http://localhost:${port} [${
          dev ? "development" : "production"
        } mode]`,
      )
    })
  } catch (error) {
    console.error("Error starting server:", error)
    process.exit(1)
  }
}

startServer()
