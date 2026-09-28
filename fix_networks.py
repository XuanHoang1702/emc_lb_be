import re

with open('docker/docker-compose.yml', 'r') as f:
    lines = f.readlines()

out = []
i = 0
while i < len(lines):
    line = lines[i]
    if line.startswith('  emc_api_node_') or line.startswith('  emc_worker:'):
        service = line.strip().strip(':')
        out.append(line)
        i += 1
        while i < len(lines) and not lines[i].startswith('  ') and not lines[i].startswith('volumes:'):
            if lines[i].strip() == 'networks:':
                out.append(lines[i])
                i += 1
                if lines[i].strip() == '- emc-network':
                    if 'api_node' in service:
                        out.append('      - public-net\n')
                        out.append('      - internal-net\n')
                        out.append('      - monitoring-net\n')
                    else:
                        out.append('      - internal-net\n')
                        out.append('      - monitoring-net\n')
                    i += 1
                continue
            out.append(lines[i])
            i += 1
        continue
        
    elif line.startswith('  nginx:'):
        out.append(line)
        i += 1
        while i < len(lines) and not lines[i].startswith('  ') and not lines[i].startswith('volumes:'):
            if lines[i].strip() == 'ports:':
                out.append(lines[i])
                i += 1
                if lines[i].strip() == '- "80:80"':
                    out.append('      - "80:80"\n')
                    out.append('      - "443:443"\n')
                    i += 1
                continue
            if lines[i].strip() == 'networks:':
                out.append(lines[i])
                i += 1
                if lines[i].strip() == '- emc-network':
                    out.append('      - public-net\n')
                    i += 1
                continue
            out.append(lines[i])
            i += 1
        continue
        
    elif line.startswith('  prometheus:') or line.startswith('  grafana:') or line.startswith('  telegraf:') or line.startswith('  postgres_exporter:') or line.startswith('  redis_exporter:') or line.startswith('  node_exporter:'):
        out.append(line)
        i += 1
        while i < len(lines) and not (lines[i].startswith('  ') and len(lines[i])>2 and lines[i][2]!=' ') and not lines[i].startswith('volumes:'):
            # Skip public ports
            if lines[i].strip() == 'ports:':
                i += 2 # skip "ports:" and the "- port:port" line
                continue
            
            if lines[i].strip() == 'networks:':
                out.append(lines[i])
                i += 1
                if lines[i].strip() == '- emc-network':
                    out.append('      - monitoring-net\n')
                    i += 1
                continue
            out.append(lines[i])
            i += 1
        continue
    
    elif line.startswith('networks:'):
        out.append('networks:\n')
        out.append('  public-net:\n')
        out.append('    driver: bridge\n')
        out.append('  internal-net:\n')
        out.append('    driver: bridge\n')
        out.append('  monitoring-net:\n')
        out.append('    driver: bridge\n')
        break
        
    else:
        out.append(line)
        i += 1

with open('docker/docker-compose.yml', 'w') as f:
    f.writelines(out)

