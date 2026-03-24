import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

export const options = {
  scenarios: {
    slots_list: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 20,
      maxVUs: 100,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],
    http_req_duration: ['p(95)<200'],
  },
};

const baseURL = __ENV.BASE_URL;
const token = __ENV.TOKEN;
const roomID = __ENV.ROOM_ID;
const date = __ENV.DATE || new Date(Date.now() + 24 * 3600 * 1000).toISOString().slice(0, 10);
const debug = __ENV.DEBUG === '1';

export default function () {
  const url = `${baseURL}/rooms/${roomID}/slots/list?date=${date}`;
  const response = http.get(url, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (debug && exec.scenario.iterationInTest === 0) {
    console.log(`url=${url}`);
    console.log(`status=${response.status}`);
    console.log(`body=${response.body}`);
  }

  check(response, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(0.1);
}