package mapper

import (
	"log"

	"github.com/hovercross/fmpxml-to-json/pkg/fmpxmlresult"
)

// Here are all the functions that read from the incoming channels, and then call their individual handlers

func (m *Mapper) handleIncomingErrorCodes() error {
	for errorCode := range m.incomingErrorCodes {
		if err := m.handleIncomingErrorCode(errorCode); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) handleIncomingProducts() error {
	for product := range m.incomingProducts {
		if err := m.handleIncomingProduct(product); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) handleIncomingDatabases() error {
	for database := range m.incomingDatabases {
		if err := m.handleIncomingDatabase(database); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) handleIncomingFields() error {
	for field := range m.incomingFields {
		if err := m.handleIncomingField(field); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) handleIncomingRows() error {
	heldRows := []fmpxmlresult.NormalizedRow{}

	// This holding loop ensures two things: First, it gives a sane default if the required data is below the rows,
	// and second it provides a facility to handle race conditions, since we can't guarantee the order of processing
	// in independent goroutines
initialLoop:
	for {
		select {
		case row, ok := <-m.incomingRows:
			// If we've closed out all our rows before handling everything else, we need to stop holding and let the chips fall where they may
			if !ok {
				break initialLoop
			}

			heldRows = append(heldRows, row)
		case <-m.startProcessingRows:
			break initialLoop // Once we've tripped the start flag, break out of this holding loop
		}
	}

	log.Println("Broke out of holding loop")

	// Process all the held rows
	for _, row := range heldRows {
		if err := m.handleIncomingRow(row); err != nil {
			return err
		}
	}

	log.Println("Finished processing held rows")

	for row := range m.incomingRows {
		if err := m.handleIncomingRow(row); err != nil {
			return err
		}
	}

	log.Println("Finished processing incoming rows")

	return nil
}

func (m *Mapper) handleIncomingMetadataEndSignals() error {
	for range m.incomingMetdataEndSignals {
		if err := m.handleIncomingMetadataEndSignal(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Mapper) handleIncomingResultSetEndSignals() error {
	for range m.incomingResultSetEndSignals {
		if err := m.handleIncomingResultSetEndSignal(); err != nil {
			return err
		}
	}

	return nil
}
