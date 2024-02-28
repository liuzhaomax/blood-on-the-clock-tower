const express = require("express")
const path = require("path")
const ejs = require("ejs")
const history = require("connect-history-api-fallback")
const favicon = require("serve-favicon")
const logger = require("morgan")
const debug = require("debug")("my-application")

const app = express()
app.disable("x-powered-by")

app.engine("html", ejs.renderFile)
    .set("views", path.resolve(__dirname, "./build"))
    .set("view engine", "html")
    .use(favicon(path.resolve(__dirname, "./build/favicon.ico")))
    .use(history({index: "/", verbose: false}))
    .use(express.static(path.resolve(__dirname, "./build")))
    // .use(logger('dev'))
    .use(logger(":method :status - :remote-addr - :date[iso] - :url - :response-time ms - :res[content-length]"))
    .use((err, req, res, next) => {
        res.locals.message = err.message
        res.locals.error = req.app.get("env") === "development" ? err : {}
        res.status(err.status || 500)
        res.render("error")
    })

const router  = express.Router()

router.get("/", (req, res, next) => {
    res.render("./build/index.html")
})

app.use("/", router)

app.listen(62120, (req, res) => {
    console.log("Server fe的be running on 62120.")
    debug()
})