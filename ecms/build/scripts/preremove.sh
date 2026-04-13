if [ $1 -eq 0 ] ; then
        # Package removal, not upgrade
        systemctl --no-reload disable ecms.service > /dev/null 2>&1 || :
        systemctl stop ecms.service > /dev/null 2>&1 || :
fi
