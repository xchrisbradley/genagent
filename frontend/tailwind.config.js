export default {
  content: ["./index.html", "./src/**/*.{html,js,ts,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [import("@tailwindcss/typography")],
};
