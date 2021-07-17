package com.example.demo;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Random;


@RestController
public class Controller {
    Logger logger = LoggerFactory.getLogger(Controller.class);

    @Autowired
	private KafkaTemplate<Object, Object> template; 

    @GetMapping("/send")
    public String sendMessageToKafka() {
        Integer requestId= new Random().nextInt(10000);
        this.template.send("a_messages", "test message with id="+requestId+" from service-c");
        String result="Sent message with id="+requestId+" from service-c";
        logger.info(result);
        return result;
    }
    
}
