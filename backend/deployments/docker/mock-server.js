const http = require('http');

const server = http.createServer((req, res) => {
    let body = '';
    req.on('data', chunk => {
        body += chunk.toString();
    });
    
    req.on('end', () => {
        console.log(`${req.method} ${req.url}`);
        res.setHeader('Content-Type', 'application/json');
        
        if (req.url === '/auth/validate') {
            // Cek token dari header Authorization
            const authHeader = req.headers['authorization'] || '';
            let role = "admin"; // Default
            
            // Jika token mengandung kata "courier", berikan role courier
            if (authHeader.toLowerCase().includes('courier')) {
                role = "courier";
            }
            
            res.writeHead(200);
            res.end(JSON.stringify({
                user_id: "mock-user-1",
                role: role
            }));
            return;
        }

        // Untuk rute lain (Tracking Service & Order Service)
        // Order Service GET /orders/:id/validate -> 200 OK
        // Tracking Service POST /events -> 200 OK
        res.writeHead(200);
        res.end(JSON.stringify({ status: "mock success" }));
    });
});

server.listen(8080, () => {
    console.log("Mock services running on port 8080");
});
