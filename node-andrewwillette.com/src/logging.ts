// import winston from "winston";
// import {production} from "./config"
//
export {logger}

function logger(data: string) {
   console.log(data)
}
//
// const level = !production ? 'debug' : 'error'
//
// const logger = winston.createLogger({
//     level: level,
//     format: winston.format.json(),
//     transports: [
//         new winston.transports.Console({})
//     ]
// });
