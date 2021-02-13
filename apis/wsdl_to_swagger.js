var apiconnWsdl = require("apiconnect-wsdl");
var yaml        = require("js-yaml");
var fs          = require("fs");

var promise = apiconnWsdl.getJsonForWSDL("Clac3.xml");

promise.then(function(wsdls){
    // Get Services from all parsed WSDLs
    var serviceData = apiconnWsdl.getWSDLServices(wsdls);

    // Loop through all services and genereate yaml file
    for (var  item in serviceData.services) {
        var serviceName = serviceData.services[item].service;
        var wsdlId = serviceData.services[item].filename;
        var wsdlEntry = apiconnWsdl.findWSDLForServiceName(wsdls, serviceName);
        var swagger = apiconnWsdl.getSwaggerForService(wsdlEntry, serviceName, wsdlId);
        var dumped = yaml.dump(swagger);

        fs.writeFile(serviceName+".yaml", dumped, (err) => {
            if(err) {
                console.log(err);
            } else {
                console.log("File written successfully");
            }
        });
    }
}, function (error) {
    console.log(error.message)
});
