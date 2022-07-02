const express = require("express");
const path = require('path')

const app = express(); // create express app
const port = 80;

app.use(express.static("build"));

app.use('*', (req, res) => {
    res.sendFile(path.join(__dirname, '/build/index.html'));
});

var server = require('http').createServer(app);
server.listen(port, "0.0.0.0");

