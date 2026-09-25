#!/bin/bash
sed -i 's|"emc_lb/src/internal/repository"|"emc_lb/src/internal/repository"\n\t"emc_lb/src/pkg/config"|g' src/internal/service/order_service.go
