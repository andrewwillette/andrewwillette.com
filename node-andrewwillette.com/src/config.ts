export {production};

let environment : String = String(process.env.REACT_APP_ENV);

let production: boolean = environment !== "dev";
