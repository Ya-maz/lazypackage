// script.js
const messages = [
    "🚀 Starting process...",
    "⏳ Doing some work...",
    "💡 Still working...",
    "✅ Done!"
];

let i = 0;
const interval = setInterval(() => {
    console.log(messages[i]);
    i++;
    if (i === messages.length) {
        clearInterval(interval);
        process.exit(0);
    }
}, 1000);

