var path = require("path");
var express = require("express");

var DIST_DIR = path.join(__dirname, "dist");
var PORT = 3000;
var app = express();

app.use(express.static(DIST_DIR));

app.get("*", function (req, res) {
  console.log(req.headers.host+" -> "+req.originalUrl);
  res.sendFile(path.join(DIST_DIR, "index.html"));
});

console.log("\nServing on serving on port "+PORT+"..");
app.listen(PORT);