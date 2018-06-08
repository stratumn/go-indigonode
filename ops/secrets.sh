# This file is sourced in launching shell to make it easier to launch another
# shell with configured ssh-agent sourcing aws_ssh_agent.sh

key_path="/keybase/team/stratumn_eng/alice_test_network/alice-test-key.pem"

export AWS_SECRET_ACCESS_KEY=ae/3o1kCQzaWpxDWeUe5vHYbComiq39weOuS/gGd
export AWS_ACCESS_KEY_ID=AKIAIY7D2TJU2B7S3STA

# if [ `basename $SHELL` = "zsh" ]; then
#     unsetopt noclobber
# fi

ldir=`dirname $0`
ssh-agent > $ldir/aws_ssh_agent.sh
source $ldir/aws_ssh_agent.sh
ssh-add $key_path
