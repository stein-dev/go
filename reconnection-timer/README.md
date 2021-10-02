**Reconnection Timer for OHP and SSHPLUS Proxy**

    Usage:  
    
    Reconnection Timer by @pigscanfly | Version: 0.0.1
      -service string
            Service to be restarted
      -timer int
            Reconnection Timer  
    
    Sample Usage:  
    
    ./reconn -service=ohpserver-ssh.service -timer=55      

### **Setup reconnection-timer as a service**   
    wget https://raw.githubusercontent.com/stein-dev/reconnection-timer/main/setup-timer-ohp.sh
    chmod 755 setup-timer-ohp.sh
    ./setup-timer-ohp.sh
    
### **Setup ohp+privoxy for ssh** 
    wget https://raw.githubusercontent.com/stein-dev/reconnection-timer/main/setup-ohp-ssh.sh
    chmod 755 setup-ohp-ssh.sh
    ./setup-ohp-ssh.sh   

### **Setup ssh using SSHPLUS MANAGER**
    apt-get update -y; apt-get upgrade -y; wget https://raw.githubusercontent.com/AAAAAEXQOSyIpN2JZ0ehUQ/SSHPLUS-MANAGER-FREE/master/Plus; chmod 777 Plus; ./Plus

### **TODO**
    Add support for SSHPLUS Proxy

### **Credits**
    @AAAAAEXQOSyIpN2JZ0ehUQ for SSHPLUS
    @lfasmpao for OHP
