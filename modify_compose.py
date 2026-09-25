import re

with open('docker/docker-compose.yml', 'r') as f:
    content = f.read()

# 1. Update networks definition at the end
content = re.sub(
    r'networks:\n  emc-network:\n    driver: bridge',
    'networks:\n  public-net:\n    driver: bridge\n  internal-net:\n    driver: bridge\n  monitoring-net:\n    driver: bridge',
    content
)

# 2. Remove ports for DBs and Monitoring
content = re.sub(r'    ports:\n      - "127\.0\.0\.1:(5432|27018|6380|4566):[0-9]+"\n', '', content)
content = re.sub(r'    ports:\n      - "(9090:9090|3000:3000|9273:9273)"\n', '', content)

# 3. Change nginx ports
content = re.sub(
    r'    ports:\n      - "80:80"\n',
    '    ports:\n      - "80:80"\n      - "443:443"\n',
    content
)

# 4. Update networks for each service type
# api_nodes -> public-net, internal-net, monitoring-net
# worker -> internal-net, monitoring-net
# postgres, mongo, redis -> internal-net, monitoring-net
# localstack -> internal-net
# nginx -> public-net
# prometheus, grafana, telegraf, postgres_exporter, redis_exporter, node_exporter -> monitoring-net

def replace_network(match):
    service = match.group(1)
    networks = []
    if 'api_node' in service:
        networks = ['public-net', 'internal-net', 'monitoring-net']
    elif service in ['emc_worker', 'postgres', 'mongo', 'redis']:
        networks = ['internal-net', 'monitoring-net']
    elif service == 'localstack':
        networks = ['internal-net']
    elif service == 'nginx':
        networks = ['public-net']
    else: # monitoring tools
        networks = ['monitoring-net']
    
    net_str = '\n'.join(f'      - {n}' for n in networks)
    return f'  {service}:\n{match.group(2)}    networks:\n{net_str}\n'

# Find service blocks and replace their networks
content = re.sub(
    r'  ([a-zA-Z0-9_]+):\n(.*?)(?:    networks:\n      - emc-network\n)',
    replace_network,
    content,
    flags=re.DOTALL
)

with open('docker/docker-compose.yml', 'w') as f:
    f.write(content)
