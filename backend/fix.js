const fs = require('fs');
let txt = fs.readFileSync('G:/sismoApp/backend/funvisis_history.json', 'utf8');
let badIdx = txt.indexOf(']":');
if (badIdx !== -1) {
    txt = txt.substring(0, badIdx + 1);
    fs.writeFileSync('G:/sismoApp/backend/funvisis_history.json', txt);
    console.log('Fixed corruption at', badIdx);
} else {
    console.log('No corruption found');
}
