CREATE TEMP TABLE "tmp_plaid_mapping" (
  "legacy_category_id" TEXT NOT NULL,
  "legacy_category" TEXT[] NOT NULL,
  "primary" TEXT NOT NULL,
  "detailed" TEXT NOT NULL
);

INSERT INTO "tmp_plaid_mapping" ("legacy_category_id", "legacy_category", "primary", "detailed")
-- Extracted from https://plaid.com/documents/transactions-personal-finance-category-mapping.json
SELECT legacy_category_id, legacy_category, possible_pfcs[0]['primary'], possible_pfcs[0]['detailed'] FROM jsonb_to_recordset($$[
  {
    "legacy_category": [
      "Bank Fees"
    ],
    "legacy_category_id": "10000000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Overdraft"
    ],
    "legacy_category_id": "10001000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OVERDRAFT_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "ATM"
    ],
    "legacy_category_id": "10002000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_ATM_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Late Payment"
    ],
    "legacy_category_id": "10003000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Fraud Dispute"
    ],
    "legacy_category_id": "10004000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Foreign Transaction"
    ],
    "legacy_category_id": "10005000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_FOREIGN_TRANSACTION_FEES"
      },
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Wire Transfer"
    ],
    "legacy_category_id": "10006000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Insufficient Funds"
    ],
    "legacy_category_id": "10007000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_INSUFFICIENT_FUNDS"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Cash Advance"
    ],
    "legacy_category_id": "10008000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Bank Fees",
      "Excess Activity"
    ],
    "legacy_category_id": "10009000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Cash Advance"
    ],
    "legacy_category_id": "11000000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_CASH_ADVANCES_AND_LOANS"
      }
    ]
  },
  {
    "legacy_category": [
      "Community"
    ],
    "legacy_category_id": "12000000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Assisted Living Services"
    ],
    "legacy_category_id": "12002000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Courts"
    ],
    "legacy_category_id": "12004000",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_GOVERNMENT_DEPARTMENTS_AND_AGENCIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Day Care and Preschools"
    ],
    "legacy_category_id": "12005000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_CHILDCARE"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Education"
    ],
    "legacy_category_id": "12008000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_EDUCATION"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Education",
      "Primary and Secondary Schools"
    ],
    "legacy_category_id": "12008003",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_EDUCATION"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Education",
      "Colleges and Universities"
    ],
    "legacy_category_id": "12008009",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_EDUCATION"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Government Departments and Agencies"
    ],
    "legacy_category_id": "12009000",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_GOVERNMENT_DEPARTMENTS_AND_AGENCIES"
      },
      {
        "primary": "INCOME",
        "detailed": "INCOME_OTHER_INCOME"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Law Enforcement"
    ],
    "legacy_category_id": "12012000",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_GOVERNMENT_DEPARTMENTS_AND_AGENCIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Libraries"
    ],
    "legacy_category_id": "12013000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_OTHER_ENTERTAINMENT"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_BOOKSTORES_AND_NEWSSTANDS"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Organizations and Associations"
    ],
    "legacy_category_id": "12015000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Organizations and Associations",
      "Youth Organizations"
    ],
    "legacy_category_id": "12015001",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_OTHER_GOVERNMENT_AND_NON_PROFIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Organizations and Associations",
      "Charities and Non-Profits"
    ],
    "legacy_category_id": "12015003",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_DONATIONS"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Post Offices"
    ],
    "legacy_category_id": "12016000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_POSTAGE_AND_SHIPPING"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Public and Social Services"
    ],
    "legacy_category_id": "12017000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_OTHER_MEDICAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Religious"
    ],
    "legacy_category_id": "12018000",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_DONATIONS"
      }
    ]
  },
  {
    "legacy_category": [
      "Community",
      "Religious",
      "Churches"
    ],
    "legacy_category_id": "12018004",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_DONATIONS"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink"
    ],
    "legacy_category_id": "13000000",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_VENDING_MACHINES"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink",
      "Bar"
    ],
    "legacy_category_id": "13001000",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_RESTAURANT"
      },
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_BEER_WINE_AND_LIQUOR"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink",
      "Restaurants"
    ],
    "legacy_category_id": "13005000",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_RESTAURANT"
      },
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_FAST_FOOD"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink",
      "Restaurants",
      "Pizza"
    ],
    "legacy_category_id": "13005012",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_FAST_FOOD"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink",
      "Restaurants",
      "Fast Food"
    ],
    "legacy_category_id": "13005032",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_FAST_FOOD"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink",
      "Restaurants",
      "Dessert"
    ],
    "legacy_category_id": "13005039",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_OTHER_FOOD_AND_DRINK"
      }
    ]
  },
  {
    "legacy_category": [
      "Food and Drink",
      "Restaurants",
      "Coffee Shop"
    ],
    "legacy_category_id": "13005043",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_COFFEE"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare"
    ],
    "legacy_category_id": "14000000",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_OTHER_MEDICAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare",
      "Healthcare Services"
    ],
    "legacy_category_id": "14001000",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_OTHER_MEDICAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare",
      "Healthcare Services",
      "Optometrists"
    ],
    "legacy_category_id": "14001005",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_EYE_CARE"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare",
      "Healthcare Services",
      "Mental Health"
    ],
    "legacy_category_id": "14001008",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_OTHER_MEDICAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare",
      "Healthcare Services",
      "Medical Supplies and Labs"
    ],
    "legacy_category_id": "14001009",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_OTHER_MEDICAL"
      },
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_PRIMARY_CARE"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare",
      "Healthcare Services",
      "Hospitals, Clinics and Medical Centers"
    ],
    "legacy_category_id": "14001010",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_PRIMARY_CARE"
      }
    ]
  },
  {
    "legacy_category": [
      "Healthcare",
      "Healthcare Services",
      "Dentists"
    ],
    "legacy_category_id": "14001012",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_DENTAL_CARE"
      }
    ]
  },
  {
    "legacy_category": [
      "Interest",
      "Interest Earned"
    ],
    "legacy_category_id": "15001000",
    "possible_pfcs": [
      {
        "primary": "INCOME",
        "detailed": "INCOME_INTEREST_EARNED"
      },
      {
        "primary": "INCOME",
        "detailed": "INCOME_DIVIDENDS"
      }
    ]
  },
  {
    "legacy_category": [
      "Interest",
      "Interest Charged"
    ],
    "legacy_category_id": "15002000",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_INTEREST_CHARGE"
      }
    ]
  },
  {
    "legacy_category": [
      "Payment"
    ],
    "legacy_category_id": "16000000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_CREDIT_CARD_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Payment",
      "Credit Card"
    ],
    "legacy_category_id": "16001000",
    "possible_pfcs": [
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_CREDIT_CARD_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Payment",
      "Rent"
    ],
    "legacy_category_id": "16002000",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_RENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Payment",
      "Loan"
    ],
    "legacy_category_id": "16003000",
    "possible_pfcs": [
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation"
    ],
    "legacy_category_id": "17000000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_OTHER_ENTERTAINMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Arts and Entertainment"
    ],
    "legacy_category_id": "17001000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_CASINOS_AND_GAMBLING"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_OTHER_ENTERTAINMENT"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Arts and Entertainment",
      "Music and Show Venues"
    ],
    "legacy_category_id": "17001007",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_OTHER_ENTERTAINMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Arts and Entertainment",
      "Movie Theatres"
    ],
    "legacy_category_id": "17001009",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Arts and Entertainment",
      "Casinos and Gaming"
    ],
    "legacy_category_id": "17001014",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_CASINOS_AND_GAMBLING"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Golf"
    ],
    "legacy_category_id": "17015000",
    "possible_pfcs": [
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_GYMS_AND_FITNESS_CENTERS"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Gyms and Fitness Centers"
    ],
    "legacy_category_id": "17018000",
    "possible_pfcs": [
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_GYMS_AND_FITNESS_CENTERS"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Sports and Recreation Camps"
    ],
    "legacy_category_id": "17041000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Sports Clubs"
    ],
    "legacy_category_id": "17042000",
    "possible_pfcs": [
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_GYMS_AND_FITNESS_CENTERS"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Stadiums and Arenas"
    ],
    "legacy_category_id": "17043000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS"
      },
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_GYMS_AND_FITNESS_CENTERS"
      }
    ]
  },
  {
    "legacy_category": [
      "Recreation",
      "Zoo"
    ],
    "legacy_category_id": "17048000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_SPORTING_EVENTS_AMUSEMENT_PARKS_AND_MUSEUMS"
      }
    ]
  },
  {
    "legacy_category": [
      "Service"
    ],
    "legacy_category_id": "18000000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Advertising and Marketing"
    ],
    "legacy_category_id": "18001000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ONLINE_MARKETPLACES"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Advertising and Marketing",
      "Print, TV, Radio and Outdoor Advertising"
    ],
    "legacy_category_id": "18001005",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Advertising and Marketing",
      "Online Advertising"
    ],
    "legacy_category_id": "18001006",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Advertising and Marketing",
      "Direct Mail and Email Marketing Services"
    ],
    "legacy_category_id": "18001008",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Advertising and Marketing",
      "Advertising Agencies and Media Buyers"
    ],
    "legacy_category_id": "18001010",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Automotive"
    ],
    "legacy_category_id": "18006000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Automotive",
      "Towing"
    ],
    "legacy_category_id": "18006001",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Automotive",
      "Maintenance and Repair"
    ],
    "legacy_category_id": "18006003",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Automotive",
      "Car Wash and Detail"
    ],
    "legacy_category_id": "18006004",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Automotive",
      "Auto Tires"
    ],
    "legacy_category_id": "18006007",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Business and Strategy Consulting"
    ],
    "legacy_category_id": "18007000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Business Services"
    ],
    "legacy_category_id": "18008000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Business Services",
      "Printing and Publishing"
    ],
    "legacy_category_id": "18008001",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Cable"
    ],
    "legacy_category_id": "18009000",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_INTERNET_AND_CABLE"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Cleaning"
    ],
    "legacy_category_id": "18011000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Computers"
    ],
    "legacy_category_id": "18012000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ELECTRONICS"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Computers",
      "Software Development"
    ],
    "legacy_category_id": "18012002",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Credit Counseling and Bankruptcy Services"
    ],
    "legacy_category_id": "18014000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Employment Agencies"
    ],
    "legacy_category_id": "18016000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_CONSULTING_AND_LEGAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Entertainment"
    ],
    "legacy_category_id": "18018000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_VIDEO_GAMES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial"
    ],
    "legacy_category_id": "18020000",
    "possible_pfcs": [
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_CASH_ADVANCES_AND_LOANS"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "Taxes"
    ],
    "legacy_category_id": "18020001",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "Stock Brokers"
    ],
    "legacy_category_id": "18020003",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_INVESTMENT_AND_RETIREMENT_FUNDS"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "Loans and Mortgages"
    ],
    "legacy_category_id": "18020004",
    "possible_pfcs": [
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_CASH_ADVANCES_AND_LOANS"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_CREDIT_CARD_PAYMENT"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_MORTGAGE_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "Financial Planning and Investments"
    ],
    "legacy_category_id": "18020007",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_INVESTMENT_AND_RETIREMENT_FUNDS"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "Banking and Finance"
    ],
    "legacy_category_id": "18020012",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_CREDIT_CARD_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "ATMs"
    ],
    "legacy_category_id": "18020013",
    "possible_pfcs": [
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_ATM_FEES"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_WITHDRAWAL"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_DEPOSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Financial",
      "Accounting and Bookkeeping"
    ],
    "legacy_category_id": "18020014",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Food and Beverage"
    ],
    "legacy_category_id": "18021000",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_RESTAURANT"
      },
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_FAST_FOOD"
      },
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_GROCERIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Food and Beverage",
      "Distribution"
    ],
    "legacy_category_id": "18021001",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_OTHER_FOOD_AND_DRINK"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Food and Beverage",
      "Catering"
    ],
    "legacy_category_id": "18021002",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_OTHER_FOOD_AND_DRINK"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Funeral Services"
    ],
    "legacy_category_id": "18022000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement"
    ],
    "legacy_category_id": "18024000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      },
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_REPAIR_AND_MAINTENANCE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Swimming Pool Maintenance and Services"
    ],
    "legacy_category_id": "18024003",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_REPAIR_AND_MAINTENANCE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Plumbing"
    ],
    "legacy_category_id": "18024007",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_REPAIR_AND_MAINTENANCE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Pest Control"
    ],
    "legacy_category_id": "18024008",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_REPAIR_AND_MAINTENANCE"
      },
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Painting"
    ],
    "legacy_category_id": "18024009",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_REPAIR_AND_MAINTENANCE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Movers"
    ],
    "legacy_category_id": "18024010",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Landscaping and Gardeners"
    ],
    "legacy_category_id": "18024013",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Home Appliances"
    ],
    "legacy_category_id": "18024018",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      },
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_FURNITURE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Home Improvement",
      "Contractors"
    ],
    "legacy_category_id": "18024024",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Household"
    ],
    "legacy_category_id": "18025000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Insurance"
    ],
    "legacy_category_id": "18030000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_INSURANCE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Internet Services"
    ],
    "legacy_category_id": "18031000",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_INTERNET_AND_CABLE"
      },
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_FLIGHTS"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Legal"
    ],
    "legacy_category_id": "18033000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_CONSULTING_AND_LEGAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Management"
    ],
    "legacy_category_id": "18036000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Manufacturing"
    ],
    "legacy_category_id": "18037000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Media Production"
    ],
    "legacy_category_id": "18038000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Oil and Gas"
    ],
    "legacy_category_id": "18042000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      },
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_GAS"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Personal Care"
    ],
    "legacy_category_id": "18045000",
    "possible_pfcs": [
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_HAIR_AND_BEAUTY"
      },
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_LAUNDRY_AND_DRY_CLEANING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Photography"
    ],
    "legacy_category_id": "18047000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Rail"
    ],
    "legacy_category_id": "18049000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_PUBLIC_TRANSIT"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Real Estate"
    ],
    "legacy_category_id": "18050000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Real Estate",
      "Real Estate Development and Title Companies"
    ],
    "legacy_category_id": "18050001",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Real Estate",
      "Real Estate Agents"
    ],
    "legacy_category_id": "18050003",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Real Estate",
      "Property Management"
    ],
    "legacy_category_id": "18050004",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Repair Services"
    ],
    "legacy_category_id": "18053000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Security and Safety"
    ],
    "legacy_category_id": "18057000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_SECURITY"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Shipping and Freight"
    ],
    "legacy_category_id": "18058000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_POSTAGE_AND_SHIPPING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Storage"
    ],
    "legacy_category_id": "18060000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_STORAGE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Subscription"
    ],
    "legacy_category_id": "18061000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_MUSIC_AND_AUDIO"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Tailors"
    ],
    "legacy_category_id": "18062000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_REPAIR_AND_MAINTENANCE"
      },
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_LAUNDRY_AND_DRY_CLEANING"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Telecommunication Services"
    ],
    "legacy_category_id": "18063000",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_TELEPHONE"
      },
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_INTERNET_AND_CABLE"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Travel Agents and Tour Operators"
    ],
    "legacy_category_id": "18067000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Utilities"
    ],
    "legacy_category_id": "18068000",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_OTHER_UTILITIES"
      },
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_GAS_AND_ELECTRICITY"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Utilities",
      "Water"
    ],
    "legacy_category_id": "18068001",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_WATER"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Utilities",
      "Sanitary and Waste Management"
    ],
    "legacy_category_id": "18068002",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_SEWAGE_AND_WASTE_MANAGEMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Utilities",
      "Gas"
    ],
    "legacy_category_id": "18068004",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_GAS_AND_ELECTRICITY"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Utilities",
      "Electric"
    ],
    "legacy_category_id": "18068005",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_GAS_AND_ELECTRICITY"
      }
    ]
  },
  {
    "legacy_category": [
      "Service",
      "Veterinarians"
    ],
    "legacy_category_id": "18069000",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_VETERINARY_SERVICES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops"
    ],
    "legacy_category_id": "19000000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ONLINE_MARKETPLACES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Adult"
    ],
    "legacy_category_id": "19001000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_OTHER_ENTERTAINMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Antiques"
    ],
    "legacy_category_id": "19002000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DISCOUNT_STORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Arts and Crafts"
    ],
    "legacy_category_id": "19003000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Auctions"
    ],
    "legacy_category_id": "19004000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Automotive"
    ],
    "legacy_category_id": "19005000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Automotive",
      "RVs and Motor Homes"
    ],
    "legacy_category_id": "19005003",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      },
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_FURNITURE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Automotive",
      "Car Parts and Accessories"
    ],
    "legacy_category_id": "19005006",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Automotive",
      "Car Dealers and Leasing"
    ],
    "legacy_category_id": "19005007",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Beauty Products"
    ],
    "legacy_category_id": "19006000",
    "possible_pfcs": [
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_HAIR_AND_BEAUTY"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Bicycles"
    ],
    "legacy_category_id": "19007000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_SPORTING_GOODS"
      },
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_FURNITURE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Boat Dealers"
    ],
    "legacy_category_id": "19008000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_OTHER_GENERAL_SERVICES"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Bookstores"
    ],
    "legacy_category_id": "19009000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_BOOKSTORES_AND_NEWSSTANDS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Cards and Stationery"
    ],
    "legacy_category_id": "19010000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Children"
    ],
    "legacy_category_id": "19011000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Clothing and Accessories"
    ],
    "legacy_category_id": "19012000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Clothing and Accessories",
      "Women's Store"
    ],
    "legacy_category_id": "19012001",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Clothing and Accessories",
      "Shoe Store"
    ],
    "legacy_category_id": "19012003",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_SPORTING_GOODS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Clothing and Accessories",
      "Men's Store"
    ],
    "legacy_category_id": "19012004",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Clothing and Accessories",
      "Kids' Store"
    ],
    "legacy_category_id": "19012006",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Computers and Electronics"
    ],
    "legacy_category_id": "19013000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ELECTRONICS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Computers and Electronics",
      "Video Games"
    ],
    "legacy_category_id": "19013001",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_VIDEO_GAMES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Computers and Electronics",
      "Mobile Phones"
    ],
    "legacy_category_id": "19013002",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ELECTRONICS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Computers and Electronics",
      "Cameras"
    ],
    "legacy_category_id": "19013003",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ELECTRONICS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Construction Supplies"
    ],
    "legacy_category_id": "19014000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_PERSONAL_LOAN_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Convenience Stores"
    ],
    "legacy_category_id": "19015000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CONVENIENCE_STORES"
      },
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_GAS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Costumes"
    ],
    "legacy_category_id": "19016000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Dance and Music"
    ],
    "legacy_category_id": "19017000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_MUSIC_AND_AUDIO"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Department Stores"
    ],
    "legacy_category_id": "19018000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DEPARTMENT_STORES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_SUPERSTORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Digital Purchase"
    ],
    "legacy_category_id": "19019000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ONLINE_MARKETPLACES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ELECTRONICS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Discount Stores"
    ],
    "legacy_category_id": "19020000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DISCOUNT_STORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Electrical Equipment"
    ],
    "legacy_category_id": "19021000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_ELECTRONICS"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_AUTOMOTIVE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Equipment Rental"
    ],
    "legacy_category_id": "19022000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_FURNITURE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DEPARTMENT_STORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Florists"
    ],
    "legacy_category_id": "19024000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Food and Beverage Store"
    ],
    "legacy_category_id": "19025000",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_BEER_WINE_AND_LIQUOR"
      },
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_OTHER_FOOD_AND_DRINK"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Food and Beverage Store",
      "Beer, Wine and Spirits"
    ],
    "legacy_category_id": "19025004",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_BEER_WINE_AND_LIQUOR"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Fuel Dealer"
    ],
    "legacy_category_id": "19026000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_GAS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Furniture and Home Decor"
    ],
    "legacy_category_id": "19027000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_FURNITURE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DEPARTMENT_STORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Gift and Novelty"
    ],
    "legacy_category_id": "19028000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      },
      {
        "primary": "PERSONAL_CARE",
        "detailed": "PERSONAL_CARE_HAIR_AND_BEAUTY"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Glasses and Optometrist"
    ],
    "legacy_category_id": "19029000",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_EYE_CARE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Hardware Store"
    ],
    "legacy_category_id": "19030000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_HARDWARE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Hobby and Collectibles"
    ],
    "legacy_category_id": "19031000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Industrial Supplies"
    ],
    "legacy_category_id": "19032000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DEPARTMENT_STORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Jewelry and Watches"
    ],
    "legacy_category_id": "19033000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Luggage"
    ],
    "legacy_category_id": "19034000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Marine Supplies"
    ],
    "legacy_category_id": "19035000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Music, Video and DVD"
    ],
    "legacy_category_id": "19036000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_TV_AND_MOVIES"
      },
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_MUSIC_AND_AUDIO"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Musical Instruments"
    ],
    "legacy_category_id": "19037000",
    "possible_pfcs": [
      {
        "primary": "ENTERTAINMENT",
        "detailed": "ENTERTAINMENT_MUSIC_AND_AUDIO"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Newsstands"
    ],
    "legacy_category_id": "19038000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_BOOKSTORES_AND_NEWSSTANDS"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CONVENIENCE_STORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Office Supplies"
    ],
    "legacy_category_id": "19039000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OFFICE_SUPPLIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Outlet"
    ],
    "legacy_category_id": "19040000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Pawn Shops"
    ],
    "legacy_category_id": "19041000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_OTHER_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Pets"
    ],
    "legacy_category_id": "19042000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_PET_SUPPLIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Pharmacies"
    ],
    "legacy_category_id": "19043000",
    "possible_pfcs": [
      {
        "primary": "MEDICAL",
        "detailed": "MEDICAL_PHARMACIES_AND_SUPPLEMENTS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Photos and Frames"
    ],
    "legacy_category_id": "19044000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Sporting Goods"
    ],
    "legacy_category_id": "19046000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_SPORTING_GOODS"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Supermarkets and Groceries"
    ],
    "legacy_category_id": "19047000",
    "possible_pfcs": [
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_GROCERIES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_SUPERSTORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Tobacco"
    ],
    "legacy_category_id": "19048000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_TOBACCO_AND_VAPE"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Toys"
    ],
    "legacy_category_id": "19049000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Vintage and Thrift"
    ],
    "legacy_category_id": "19050000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_DISCOUNT_STORES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_GIFTS_AND_NOVELTIES"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Warehouses and Wholesale Stores"
    ],
    "legacy_category_id": "19051000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_SUPERSTORES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Wedding and Bridal"
    ],
    "legacy_category_id": "19052000",
    "possible_pfcs": [
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_OTHER_GENERAL_MERCHANDISE"
      },
      {
        "primary": "GENERAL_MERCHANDISE",
        "detailed": "GENERAL_MERCHANDISE_CLOTHING_AND_ACCESSORIES"
      }
    ]
  },
  {
    "legacy_category": [
      "Shops",
      "Lawn and Garden"
    ],
    "legacy_category_id": "19054000",
    "possible_pfcs": [
      {
        "primary": "HOME_IMPROVEMENT",
        "detailed": "HOME_IMPROVEMENT_OTHER_HOME_IMPROVEMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Tax",
      "Refund"
    ],
    "legacy_category_id": "20001000",
    "possible_pfcs": [
      {
        "primary": "INCOME",
        "detailed": "INCOME_TAX_REFUND"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Tax",
      "Payment"
    ],
    "legacy_category_id": "20002000",
    "possible_pfcs": [
      {
        "primary": "GOVERNMENT_AND_NON_PROFIT",
        "detailed": "GOVERNMENT_AND_NON_PROFIT_TAX_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer"
    ],
    "legacy_category_id": "21000000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Internal Account Transfer"
    ],
    "legacy_category_id": "21001000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Billpay"
    ],
    "legacy_category_id": "21003000",
    "possible_pfcs": [
      {
        "primary": "RENT_AND_UTILITIES",
        "detailed": "RENT_AND_UTILITIES_OTHER_UTILITIES"
      },
      {
        "primary": "LOAN_PAYMENTS",
        "detailed": "LOAN_PAYMENTS_CREDIT_CARD_PAYMENT"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Check"
    ],
    "legacy_category_id": "21004000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Credit"
    ],
    "legacy_category_id": "21005000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Debit"
    ],
    "legacy_category_id": "21006000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Deposit"
    ],
    "legacy_category_id": "21007000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_DEPOSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Deposit",
      "Check"
    ],
    "legacy_category_id": "21007001",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_DEPOSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Deposit",
      "ATM"
    ],
    "legacy_category_id": "21007002",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_DEPOSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Keep the Change Savings Program"
    ],
    "legacy_category_id": "21008000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_SAVINGS"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_SAVINGS"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Payroll"
    ],
    "legacy_category_id": "21009000",
    "possible_pfcs": [
      {
        "primary": "INCOME",
        "detailed": "INCOME_WAGES"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Payroll",
      "Benefits"
    ],
    "legacy_category_id": "21009001",
    "possible_pfcs": [
      {
        "primary": "INCOME",
        "detailed": "INCOME_UNEMPLOYMENT"
      },
      {
        "primary": "INCOME",
        "detailed": "INCOME_RETIREMENT_PENSION"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party"
    ],
    "legacy_category_id": "21010000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Venmo"
    ],
    "legacy_category_id": "21010001",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Square Cash"
    ],
    "legacy_category_id": "21010002",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Square"
    ],
    "legacy_category_id": "21010003",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "FOOD_AND_DRINK",
        "detailed": "FOOD_AND_DRINK_OTHER_FOOD_AND_DRINK"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "PayPal"
    ],
    "legacy_category_id": "21010004",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Coinbase"
    ],
    "legacy_category_id": "21010006",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_INVESTMENT_AND_RETIREMENT_FUNDS"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Chase QuickPay"
    ],
    "legacy_category_id": "21010007",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "BANK_FEES",
        "detailed": "BANK_FEES_OTHER_BANK_FEES"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Acorns"
    ],
    "legacy_category_id": "21010008",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_INVESTMENT_AND_RETIREMENT_FUNDS"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Digit"
    ],
    "legacy_category_id": "21010009",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_SAVINGS"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_ACCOUNTING_AND_FINANCIAL_PLANNING"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Third Party",
      "Betterment"
    ],
    "legacy_category_id": "21010010",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Wire"
    ],
    "legacy_category_id": "21011000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Withdrawal"
    ],
    "legacy_category_id": "21012000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_WITHDRAWAL"
      },
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_INVESTMENT_AND_RETIREMENT_FUNDS"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Withdrawal",
      "Check"
    ],
    "legacy_category_id": "21012001",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_OTHER_TRANSFER_OUT"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Withdrawal",
      "ATM"
    ],
    "legacy_category_id": "21012002",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_WITHDRAWAL"
      }
    ]
  },
  {
    "legacy_category": [
      "Transfer",
      "Save As You Go"
    ],
    "legacy_category_id": "21013000",
    "possible_pfcs": [
      {
        "primary": "TRANSFER_OUT",
        "detailed": "TRANSFER_OUT_SAVINGS"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_SAVINGS"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel"
    ],
    "legacy_category_id": "22000000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_FLIGHTS"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Airlines and Aviation Services"
    ],
    "legacy_category_id": "22001000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_FLIGHTS"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Airports"
    ],
    "legacy_category_id": "22002000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      },
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_LODGING"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Boat"
    ],
    "legacy_category_id": "22003000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Bus Stations"
    ],
    "legacy_category_id": "22004000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      },
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_PUBLIC_TRANSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Car and Truck Rentals"
    ],
    "legacy_category_id": "22005000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_RENTAL_CARS"
      },
      {
        "primary": "GENERAL_SERVICES",
        "detailed": "GENERAL_SERVICES_POSTAGE_AND_SHIPPING"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Car Service"
    ],
    "legacy_category_id": "22006000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_TAXIS_AND_RIDE_SHARES"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Car Service",
      "Ride Share"
    ],
    "legacy_category_id": "22006001",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_TAXIS_AND_RIDE_SHARES"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Charter Buses"
    ],
    "legacy_category_id": "22007000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_PUBLIC_TRANSIT"
      },
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Cruises"
    ],
    "legacy_category_id": "22008000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Gas Stations"
    ],
    "legacy_category_id": "22009000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_GAS"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Lodging"
    ],
    "legacy_category_id": "22012000",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_LODGING"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Lodging",
      "Lodges and Vacation Rentals"
    ],
    "legacy_category_id": "22012002",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_OTHER_TRAVEL"
      },
      {
        "primary": "TRANSFER_IN",
        "detailed": "TRANSFER_IN_ACCOUNT_TRANSFER"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Lodging",
      "Hotels and Motels"
    ],
    "legacy_category_id": "22012003",
    "possible_pfcs": [
      {
        "primary": "TRAVEL",
        "detailed": "TRAVEL_LODGING"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Parking"
    ],
    "legacy_category_id": "22013000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_PARKING"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Public Transportation Services"
    ],
    "legacy_category_id": "22014000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_PUBLIC_TRANSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Rail"
    ],
    "legacy_category_id": "22015000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_PUBLIC_TRANSIT"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Taxi"
    ],
    "legacy_category_id": "22016000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_TAXIS_AND_RIDE_SHARES"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Tolls and Fees"
    ],
    "legacy_category_id": "22017000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_TOLLS"
      }
    ]
  },
  {
    "legacy_category": [
      "Travel",
      "Transportation Centers"
    ],
    "legacy_category_id": "22018000",
    "possible_pfcs": [
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_OTHER_TRANSPORTATION"
      },
      {
        "primary": "TRANSPORTATION",
        "detailed": "TRANSPORTATION_GAS"
      }
    ]
  }
]$$) as mapping(
  "legacy_category" TEXT[],
  "legacy_category_id" TEXT,
  "possible_pfcs" JSONB
);

ALTER TABLE "transactions" ADD COLUMN "category" TEXT;
UPDATE "transactions" SET "category" = "tmp_plaid_mapping"."detailed"
FROM "tmp_plaid_mapping"
WHERE "tmp_plaid_mapping"."legacy_category" = "transactions"."categories";

ALTER TABLE "plaid_transactions" ADD COLUMN "category" TEXT;
UPDATE "plaid_transactions" SET "category" = "tmp_plaid_mapping"."detailed"
FROM "tmp_plaid_mapping"
WHERE "tmp_plaid_mapping"."legacy_category" = "plaid_transactions"."categories";

DROP TABLE "tmp_plaid_mapping";
