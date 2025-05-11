const express = require('express');
const { Kafka } = require('kafkajs');
const { Client } = require('cassandra-driver');
const winston = require('winston');

// Initialize logger
const logger = winston.createLogger({
  level: 'info',
  format: winston.format.json(),
  transports: [
    new winston.transports.Console()
  ]
});

// Initialize Kafka consumer
const kafka = new Kafka({
  clientId: 'cad-validation-service',
  brokers: ['kafka:9092']
});

const consumer = kafka.consumer({ groupId: 'validation-group' });

// Initialize Cassandra client
const cassandraClient = new Client({
  contactPoints: ['cassandra'],
  localDataCenter: 'datacenter1',
  keyspace: 'cad_automation'
});

const app = express();
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

// Validation endpoint
app.post('/validate/:fileId', async (req, res) => {
  const { fileId } = req.params;
  
  try {
    // Get file metadata from Cassandra
    const result = await cassandraClient.execute(
      'SELECT * FROM files WHERE id = ?',
      [fileId],
      { prepare: true }
    );

    if (result.rowLength === 0) {
      return res.status(404).json({ error: 'File not found' });
    }

    const file = result.first();
    
    // Perform validation (simplified example)
    const validationResult = {
      isValid: true,
      issues: [],
      timestamp: new Date().toISOString()
    };

    // Update file status in Cassandra
    await cassandraClient.execute(
      'UPDATE files SET status = ?, validation_result = ? WHERE id = ?',
      ['validated', JSON.stringify(validationResult), fileId],
      { prepare: true }
    );

    res.json(validationResult);
  } catch (error) {
    logger.error('Validation error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

// Start Kafka consumer
async function startConsumer() {
  await consumer.connect();
  await consumer.subscribe({ topic: 'cad-files', fromBeginning: true });

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const fileId = message.value.toString();
      logger.info(`Processing file: ${fileId}`);

      try {
        // Trigger validation
        const response = await fetch(`http://localhost:${process.env.PORT || 3000}/validate/${fileId}`, {
          method: 'POST'
        });

        if (!response.ok) {
          throw new Error(`Validation failed: ${response.statusText}`);
        }

        logger.info(`Validation completed for file: ${fileId}`);
      } catch (error) {
        logger.error(`Error processing file ${fileId}:`, error);
      }
    },
  });
}

// Start the server
const PORT = process.env.PORT || 3000;
app.listen(PORT, async () => {
  logger.info(`Validation service listening on port ${PORT}`);
  await startConsumer();
}); 