# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    config.vm.define "client" do |client|
        client.vm.box = "geerlingguy/centos6"
        client.vm.network "private_network", ip: "192.168.50.4"
    end

    config.vm.define "target1" do |target1|
        target1.vm.box = "geerlingguy/centos6"
        target1.vm.network "private_network", ip: "192.168.50.5"
    end

    config.vm.define "target2" do |target2|
        target2.vm.box = "geerlingguy/centos6"
        target2.vm.network "private_network", ip: "192.168.50.6"
    end
end
