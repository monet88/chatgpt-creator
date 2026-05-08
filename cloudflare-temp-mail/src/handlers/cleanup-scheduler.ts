import { runCleanup } from '../services/cleanup-service';
import type { Env } from '../types';

export const handleScheduled = async (_controller: ScheduledController, env: Env) => {
  await runCleanup(env);
};
