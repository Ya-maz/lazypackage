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
  }
}, 1000);

