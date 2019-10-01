var path = require("path");
var express = require("express");

var DIST_DIR = path.join(__dirname, "dist");
var PORT = 3000;
var app = express();

app.use((req, res, next) => {
  res.set("Cache-Control", "no-cache, no-store, must-revalidate");
  res.set("Pragma", "no-cache");
  res.set("Expires", "0");
  next();
});
app.use(express.static(DIST_DIR));

app.get("*", function (req, res) {
  console.log(req.headers.host + "-" + req.method + " -> "+req.originalUrl);
  res.sendFile(path.join(DIST_DIR, "index.html"));
});

app.all("*", function (req, res) {
  console.log(req.headers.host + "-" + req.method + " -> "+req.originalUrl);
  res.status(404);
  res.json({});
});

console.log("\nServing on serving on port "+PORT+"..");
app.listen(PORT);