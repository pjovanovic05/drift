# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    config.vm.define "client" do |client|
        client.vm.box = "ubuntu/xenial64"
        client.vm.network "private_network", ip: "192.168.50.4"
    end

    config.vm.define "target1" do |target1|
        target1.vm.box = "ubuntu/xenial64"
        target1.vm.network "private_network", ip: "192.168.50.5"
    end

    config.vm.define "target2" do |target2|
        target2.vm.box = "ubuntu/xenial64"
        target2.vm.network "private_network", ip: "192.168.50.6"
    end

  # config.vm.provider "virtualbox" do |vb|
  #   # Display the VirtualBox GUI when booting the machine
  #   vb.gui = true
  #
  #   # Customize the amount of memory on the VM:
  #   vb.memory = "1024"
  # end
  #
  # View the documentation for the provider you are using for more
  # information on available options.

  # Enable provisioning with a shell script. Additional provisioners such as
  # Puppet, Chef, Ansible, Salt, and Docker are also available. Please see the
  # documentation for more information about their specific syntax and use.
  # config.vm.provision "shell", inline: <<-SHELL
  #   apt-get update
  #   apt-get install -y apache2
  # SHELL
end
