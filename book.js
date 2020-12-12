const path = require("path")
const fs = require("fs")
module.exports = {
    "root": "./docs",
    "title": "つくって学ぶKubebuilder",
    "plugins": [
        "include-codeblock"
    ],
    "pluginsConfig": {
        "include-codeblock": {
            "template": path.join(__dirname,"template.hbs")
        }
    }
};
