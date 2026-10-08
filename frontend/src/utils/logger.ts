import { createConsola } from 'consola';

export const logger = createConsola({
  // Log level (0: fatal, 1: error, 2: warn, 3: log/info, 4: debug, 5: trace)
  level: import.meta.env.MODE === 'development' ? 4 : 3,
  // Add a nice fancy format
  formatOptions: {
    colors: true,
    compact: false,
    date: true,
  }
});
