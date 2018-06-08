# HELLO !

To deploy an alice test network on AWS:

```
# install python dependencies
pip install -r requirements.txt

# update ansible dependencies
ansible-galaxy install -r ansible/requirements.yml -p ansible/roles/

source secrets.sh

./launch_network.sh
```
