import env from "../env/env.json"

const config = {
    domain: env.prod.fe.domain,
    beBaseUrl: env.run_mode === "prod" ?
        `${env.prod.be.protocol}://${env.prod.be.domain}` :
        `${env.dev.be.protocol}://${env.dev.be.host}:${env.dev.be.port}`
}
export default config