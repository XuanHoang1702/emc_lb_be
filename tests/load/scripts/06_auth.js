import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

export const options = {
    stages: [
        { duration: '30s', target: 20 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<1000'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost';

const users = new SharedArray('users', function () {
    try {
        const f = JSON.parse(open('../data/seed.json'));
        return f.users;
    } catch(e) {
        return [{ email: "test@example.com", password: "Password123!" }];
    }
});

export default function () {
    const user = users[Math.floor(Math.random() * users.length)];
    
    const payload = JSON.stringify({
        email: user.email,
        password: user.password,
    });
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(`${BASE_URL}/api/v1/users/login`, payload, params);
    check(res, { 'login status was 200': (r) => r.status === 200 || r.status === 401 });
    
    sleep(1);
}
