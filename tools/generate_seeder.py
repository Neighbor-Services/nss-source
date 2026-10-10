import json
import zipfile
import xml.etree.ElementTree as ET

categories_data = [
    {
        "name": "Home Repairs and Improvements",
        "description": "Professional residential and commercial repairs, plumbing, electrical, HVAC, carpentry, painting, roofing, and remodeling services.",
        "image": "categories/home_repairs.jpg",
        "services": [
            {
                "name": "Handyman Services",
                "description": "General home repairs, fixture installations, drywall patching, furniture assembly, and small maintenance projects.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Drywall repair", "Furniture assembly", "TV mounting", "Fixture installation", "Door repair", "Minor home repairs"]
            },
            {
                "name": "Plumbing",
                "description": "Certified plumbing repairs, leak detection, drain clearing, water heater installation, and pipe maintenance.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Leak repair", "Drain cleaning", "Water heaters", "Fixture installation", "Pipe repair", "Toilet repair"]
            },
            {
                "name": "Electrical Services",
                "description": "Licensed electrical wiring, outlet repair, panel upgrades, lighting installations, and EV charger setup.",
                "price": 95.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Light fixture installation", "Ceiling fans", "Outlet and switch repair", "Panel upgrades", "Wiring", "EV charger installation"]
            },
            {
                "name": "Heating and Air Conditioning",
                "description": "HVAC system inspection, heating repair, AC maintenance, duct cleaning, and heat pump installations.",
                "price": 110.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["AC repair", "Furnace maintenance", "HVAC installation", "Duct cleaning", "Thermostat installation", "Heat pump service"]
            },
            {
                "name": "Appliance Repair",
                "description": "Diagnosis and repair for refrigerators, washers, dryers, dishwashers, ovens, and household appliances.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Refrigerator repair", "Washer and dryer repair", "Dishwasher repair", "Oven and stove repair", "Microwave repair"]
            },
            {
                "name": "Carpentry",
                "description": "Custom cabinetry, trim carpentry, baseboards, crown molding, framing, and wood structural repairs.",
                "price": 70.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Custom cabinetry", "Trim and molding", "Framing", "Deck repair", "Wood rot repair", "Shelving"]
            },
            {
                "name": "Painting",
                "description": "Interior and exterior house painting, cabinet refinishing, deck staining, and wallpaper removal.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Interior painting", "Exterior painting", "Cabinet painting", "Deck staining", "Drywall patching", "Wallpaper removal"]
            },
            {
                "name": "Flooring",
                "description": "Hardwood floor installation, tile and grout work, vinyl plank, laminate, and carpet laying.",
                "price": 65.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Hardwood installation", "Tile and grout", "Laminate flooring", "Carpet installation", "Floor refinishing", "Vinyl plank"]
            },
            {
                "name": "Roofing",
                "description": "Roof leak repair, shingle replacement, flashing maintenance, inspections, and full roof repairs.",
                "price": 120.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Roof leak repair", "Shingle replacement", "Roof inspections", "Flashing repair", "Gutter integration", "Flat roofing"]
            },
            {
                "name": "Remodeling",
                "description": "Full kitchen and bathroom remodels, basement finishing, room conversions, and home renovations.",
                "price": 150.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Kitchen remodeling", "Bathroom remodeling", "Basement finishing", "Room additions", "Attic conversion"]
            },
            {
                "name": "Doors and Windows",
                "description": "Window installation, glass replacement, exterior and interior door fitting, sliding doors, and screens.",
                "price": 80.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Window replacement", "Exterior door installation", "Sliding door repair", "Storm doors", "Screen replacement"]
            },
            {
                "name": "Locksmith Services",
                "description": "Emergency lockouts, rekeying locks, smart lock installations, deadbolts, and security hardware.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Emergency lockout", "Rekeying", "Smart lock installation", "Deadbolt repair", "Master key systems"]
            }
        ]
    },
    {
        "name": "Cleaning and Household Services",
        "description": "Residential home cleaning, deep cleaning, commercial janitorial, pressure washing, pest control, and junk removal.",
        "image": "categories/cleaning.jpg",
        "services": [
            {
                "name": "House Cleaning",
                "description": "Recurring maid service, deep cleaning, move-in/move-out cleans, and eco-friendly home sanitization.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Standard recurring cleaning", "Deep cleaning", "Move-in/move-out cleaning", "Eco-friendly cleaning", "Post-renovation"]
            },
            {
                "name": "Commercial Cleaning",
                "description": "Office cleaning, retail space sanitization, janitorial maintenance, and commercial facility care.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Office cleaning", "Janitorial services", "Retail cleaning", "Medical office sanitization"]
            },
            {
                "name": "Carpet and Upholstery Cleaning",
                "description": "Deep steam carpet cleaning, sofa and couch shampooing, stain removal, and pet odor treatments.",
                "price": 70.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Steam cleaning", "Stain removal", "Pet odor removal", "Sofa and couch cleaning", "Rug cleaning"]
            },
            {
                "name": "Window Cleaning",
                "description": "Streak-free interior and exterior window washing, screen scrubbing, and skylight cleaning.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Interior window washing", "Exterior window washing", "Screen cleaning", "Skylight cleaning"]
            },
            {
                "name": "Pressure Washing",
                "description": "High-pressure washing for driveways, sidewalks, home siding, patios, decks, and fences.",
                "price": 80.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Driveway cleaning", "Siding power wash", "Deck and patio wash", "Roof soft wash", "Fence wash"]
            },
            {
                "name": "Laundry and Dry Cleaning",
                "description": "Wash, dry, fold services, dry cleaning pickup and delivery, garment pressing, and ironing.",
                "price": 30.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wash and fold", "Dry cleaning pickup", "Ironing", "Garment care", "Bedding and linen"]
            },
            {
                "name": "Home Organization",
                "description": "Professional organizing for closets, pantries, garages, kitchens, and decluttering assistance.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Closet organization", "Garage cleanout", "Kitchen pantry organizing", "Decluttering assistance"]
            },
            {
                "name": "Pest Control",
                "description": "Extermination and preventative treatments for termites, rodents, bed bugs, ants, and roaches.",
                "price": 90.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Termite treatment", "Rodent control", "Bed bug extermination", "Ant and roach treatment", "Mosquito fogging"]
            },
            {
                "name": "Junk Removal",
                "description": "Eco-friendly hauling of old furniture, appliances, yard waste, construction debris, and estate cleanouts.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Furniture haul away", "Appliance disposal", "Estate cleanout", "Yard waste removal", "Construction debris"]
            }
        ]
    },
    {
        "name": "Lawn and Outdoor Services",
        "description": "Lawn mowing, landscaping design, tree trimming, irrigation systems, fences, patios, and pool care.",
        "image": "categories/lawn_outdoor.jpg",
        "services": [
            {
                "name": "Lawn Care",
                "description": "Lawn mowing, edging, weed trimming, fertilization, aeration, overseeding, and seasonal yard cleanup.",
                "price": 40.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Lawn mowing", "Edging", "Fertilization", "Aeration and overseeding", "Weed control"]
            },
            {
                "name": "Landscaping",
                "description": "Landscape design, plant installation, mulching, sod laying, garden bed edging, and rock features.",
                "price": 65.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Landscape design", "Mulching", "Sod installation", "Plant and shrub planting", "Retaining walls"]
            },
            {
                "name": "Tree Care",
                "description": "Tree trimming, branch pruning, tree removal, stump grinding, and certified arborist consultations.",
                "price": 120.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Tree trimming", "Tree removal", "Stump grinding", "Arborist consulting", "Emergency branch removal"]
            },
            {
                "name": "Irrigation and Sprinklers",
                "description": "Sprinkler system installation, zone repairs, head replacements, drip lines, and winterization.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Sprinkler system repair", "Drip irrigation", "Winterization", "Backflow testing", "Timer programming"]
            },
            {
                "name": "Fence and Gate Services",
                "description": "Wood, vinyl, and chain link fence construction, post replacements, gate repairs, and automation.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wood fence installation", "Chain link", "Vinyl fencing", "Gate repair", "Automatic gate openers"]
            },
            {
                "name": "Decks and Patios",
                "description": "Custom deck building, paver patios, concrete stamping, porch repairs, and pergola construction.",
                "price": 130.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Paver patio construction", "Deck building", "Concrete pouring", "Pergolas and gazebos"]
            },
            {
                "name": "Pool Maintenance",
                "description": "Weekly pool cleaning, chemical balancing, filter backwashing, pump repair, and equipment service.",
                "price": 80.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Chemical balancing", "Pool opening/closing", "Filter cleaning", "Leak detection", "Tile scrubbing"]
            },
            {
                "name": "Gutter Cleaning",
                "description": "Debris removal from gutters and downspouts, gutter guard installation, and leak patching.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Gutter debris removal", "Gutter guard installation", "Downspout flushing", "Minor gutter repair"]
            }
        ]
    },
    {
        "name": "Moving and Storage",
        "description": "Residential moving, commercial office moves, packing/unpacking, furniture assembly, and storage solutions.",
        "image": "categories/moving_storage.jpg",
        "services": [
            {
                "name": "Residential Moving",
                "description": "Local and long-distance home moving, apartment relocations, and professional furniture handling.",
                "price": 95.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Local home moving", "Long distance moving", "Apartment moving", "Senior relocation"]
            },
            {
                "name": "Commercial Moving",
                "description": "Office relocations, cubicle moves, corporate equipment transfer, and business transitions.",
                "price": 120.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Office relocation", "Business equipment move", "Retail store moving"]
            },
            {
                "name": "Packing and Unpacking",
                "description": "Full-service packing with boxes, bubble wrap, fragile crating, and organized unpacking support.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Full packing service", "Fragile item crating", "Box labeling", "Unpacking assistance"]
            },
            {
                "name": "Furniture Assembly",
                "description": "Assembly of IKEA flat-pack furniture, bed frames, dressers, desks, and home gym equipment.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["IKEA furniture assembly", "Bed frames", "Desks and tables", "Exercise equipment assembly"]
            },
            {
                "name": "Loading and Unloading",
                "description": "Heavy lifting labor for loading or unloading moving trucks, Pods, trailers, and storage units.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Truck loading labor", "Storage container loading", "Heavy lifting labor"]
            },
            {
                "name": "Storage Services",
                "description": "Short and long term secure storage units, climate control, vehicle storage, and portable pods.",
                "price": 60.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Secure unit storage", "Climate controlled storage", "Vehicle storage", "Mobile storage delivery"]
            }
        ]
    },
    {
        "name": "Automotive Services",
        "description": "Mobile mechanics, car detailing, mobile tire service, auto glass, battery replacement, and towing.",
        "image": "categories/automotive.jpg",
        "services": [
            {
                "name": "Auto Repair and Maintenance",
                "description": "On-site brake pad replacement, oil changes, engine diagnostic scans, starters, and alternators.",
                "price": 80.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Brake service", "Oil change", "Engine diagnostics", "Transmission repair", "Tune-ups", "Alternators"]
            },
            {
                "name": "Car Washing and Detailing",
                "description": "Mobile hand car wash, interior shampooing, exterior buffing, clay bar, and ceramic coating.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile car wash", "Interior deep detail", "Paint correction", "Ceramic coating", "Leather conditioning"]
            },
            {
                "name": "Tire Services",
                "description": "Mobile tire mounting, balancing, flat tire puncture repairs, and tire rotation at your location.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Tire mounting", "Balancing", "Flat tire repair", "Tire rotation", "Mobile tire change"]
            },
            {
                "name": "Auto Glass Services",
                "description": "Windshield chip repairs, cracked glass replacement, side window and rearview mirror replacements.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Windshield chip repair", "Windshield replacement", "Side glass", "Mirror replacement"]
            },
            {
                "name": "Battery Services",
                "description": "On-demand battery jump starts, battery diagnostic health tests, and new battery installations.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Battery jump start", "Battery testing", "On-site battery replacement", "Terminal cleaning"]
            },
            {
                "name": "Roadside Assistance",
                "description": "Emergency roadside help including lockout rescue, emergency fuel delivery, and flat tire swaps.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Lockout service", "Fuel delivery", "Winch out", "Flat tire change"]
            },
            {
                "name": "Towing",
                "description": "Emergency flatbed and wheel-lift towing for cars, SUVs, motorcycles, and disabled vehicles.",
                "price": 95.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Flatbed towing", "Wheel lift towing", "Long distance towing", "Accident recovery"]
            },
            {
                "name": "Vehicle Inspections",
                "description": "Comprehensive pre-purchase mechanical inspections, safety checks, and computer diagnostics.",
                "price": 70.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Pre-purchase inspections", "Emissions check", "Safety inspection", "Diagnostic scanning"]
            }
        ]
    },
    {
        "name": "Transportation and Delivery",
        "description": "Passenger transit, private chauffeurs, airport shuttles, couriers, freight delivery, and errand runs.",
        "image": "categories/transportation.jpg",
        "services": [
            {
                "name": "Passenger Transportation",
                "description": "Reliable point-to-point passenger rides, private transportation, and non-emergency transit.",
                "price": 35.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Local rides", "Point-to-point transit", "Group shuttle", "Non-emergency medical transit"]
            },
            {
                "name": "Chauffeur Services",
                "description": "Executive black car services, luxury sedan travel, dedicated personal drivers, and VIP transit.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Executive black car", "Luxury sedan", "Hourly private driver", "Special event transport"]
            },
            {
                "name": "Airport Transfers",
                "description": "Scheduled airport departures and pickups, luggage assistance, and flight status tracking.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Airport drop-off", "Airport pickup", "Flight tracking arrival", "Luggage assistance"]
            },
            {
                "name": "Courier Services",
                "description": "Same-day legal document deliveries, medical parcel courier, and urgent package distribution.",
                "price": 30.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Same-day legal documents", "Medical sample courier", "Package delivery", "Express parcel"]
            },
            {
                "name": "Grocery and Retail Delivery",
                "description": "Personal supermarket shopping, grocery delivery, and store merchandise pickup to your door.",
                "price": 25.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Supermarket grocery run", "Pharmacy medicine pickup", "Retail merchandise delivery"]
            },
            {
                "name": "Freight and Large Item Delivery",
                "description": "Cargo van transport, appliance delivery, pallet moving, and oversized goods transportation.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Pallet transport", "Large appliance delivery", "Furniture transport", "Cargo van transport"]
            },
            {
                "name": "Errand Services",
                "description": "Personal assistant errands, post office drops, prescription pickups, and general to-do tasks.",
                "price": 25.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Personal errands", "Post office runs", "Shopping assistance", "Dry cleaning pickup"]
            }
        ]
    },
    {
        "name": "Beauty and Personal Style",
        "description": "Hair styling, barbering, nail care, makeup artistry, facials, waxing, lash extensions, and personal styling.",
        "image": "categories/beauty_grooming.jpg",
        "services": [
            {
                "name": "Hair Services",
                "description": "Haircuts, blowouts, hair coloring, balayage, braiding, natural hair care, locs, and extensions.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Braiding", "Natural hair", "Coloring", "Extensions", "Locs", "Blowouts", "Haircuts", "Balayage"]
            },
            {
                "name": "Barbering",
                "description": "Classic fades, men haircuts, beard sculpting, line-ups, and luxury hot towel straight-razor shaves.",
                "price": 35.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Men haircuts", "Fades", "Beard sculpting", "Hot towel shave", "Edge ups", "Lineups"]
            },
            {
                "name": "Nail Care",
                "description": "Manicures, spa pedicures, acrylic full sets, gel polish, dip powder, and custom nail art.",
                "price": 40.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Manicures", "Pedicures", "Gel nails", "Acrylic full sets", "Dip powder", "Nail art"]
            },
            {
                "name": "Makeup",
                "description": "Bridal makeup, glam makeup for parties, photoshoot beauty, and special occasion styling.",
                "price": 65.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Bridal makeup", "Glam makeup", "Photoshoot makeup", "Special event makeup", "SFX makeup"]
            },
            {
                "name": "Skin Care and Facials",
                "description": "Hydrating facials, deep pore cleansing, microdermabrasion, anti-aging therapies, and acne care.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Hydrating facials", "Chemical peels", "Microdermabrasion", "Anti-aging treatments", "Acne care"]
            },
            {
                "name": "Waxing and Hair Removal",
                "description": "Full body waxing, Brazilian wax, precision eyebrow threading, sugaring, and laser hair removal.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Brazilian waxing", "Eyebrow threading", "Body waxing", "Laser hair removal", "Sugaring"]
            },
            {
                "name": "Lash and Brow Services",
                "description": "Classic and volume lash extensions, lash lifts, brow lamination, microblading, and brow tinting.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Lash extensions", "Lash lift and tint", "Brow lamination", "Microblading", "Brow shaping"]
            },
            {
                "name": "Personal Styling",
                "description": "Wardrobe consulting, color palette analysis, personal shopping, closet audits, and event styling.",
                "price": 70.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wardrobe styling", "Color analysis", "Personal shopping", "Closet audit", "Event outfit curation"]
            },
            {
                "name": "Tailoring and Dressmaking",
                "description": "Clothing alterations, hemming pants, custom suit tailoring, dress adjustments, and zipper repair.",
                "price": 35.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Alterations", "Hemming", "Custom suits", "Bridal gown alterations", "Zipper replacement"]
            }
        ]
    },
    {
        "name": "Fitness and Wellness",
        "description": "Personal fitness training, group classes, yoga, sports coaching, massage therapy, and wellness coaching.",
        "image": "categories/fitness_wellness.jpg",
        "services": [
            {
                "name": "Personal Training",
                "description": "One-on-one fitness training, strength conditioning, fat loss programs, and home workout routines.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["One-on-one training", "Weight loss", "Strength and conditioning", "Mobility training", "In-home fitness"]
            },
            {
                "name": "Fitness Classes",
                "description": "High-intensity interval training (HIIT), bootcamps, spin cycling, Zumba, and cardio fitness classes.",
                "price": 25.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["HIIT classes", "Bootcamps", "Group cycling", "Zumba", "Barre classes"]
            },
            {
                "name": "Yoga and Pilates",
                "description": "Vinyasa yoga, Hatha yoga, Mat and Reformer Pilates, meditation, and guided breathwork sessions.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Vinyasa yoga", "Hatha yoga", "Mat Pilates", "Reformer Pilates", "Breathwork", "Meditation"]
            },
            {
                "name": "Sports Coaching",
                "description": "Private and youth coaching for tennis, swimming, basketball, soccer, golf, and athletic skills.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Tennis lessons", "Swimming coaching", "Golf instruction", "Soccer skills", "Basketball training"]
            },
            {
                "name": "Massage Therapy",
                "description": "In-home deep tissue massage, Swedish relaxation, sports massage, prenatal massage, and trigger point therapy.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Deep tissue massage", "Swedish massage", "Sports massage", "Prenatal massage", "Hot stone therapy"]
            },
            {
                "name": "Wellness Coaching",
                "description": "Lifestyle wellness, holistic health guidance, stress relief coaching, and sustainable habit formation.",
                "price": 65.00,
                "location": "ONLINE",
                "specialties": ["Holistic lifestyle coaching", "Stress management", "Sleep coaching", "Habit formation"]
            }
        ]
    },
    {
        "name": "Childcare and Early Learning",
        "description": "Vetted babysitters, full-time nannies, licensed daycares, preschool programs, and after-school care.",
        "image": "categories/childcare.jpg",
        "services": [
            {
                "name": "Babysitting",
                "description": "Trusted on-demand babysitting, date-night childcare, infant care, and weekend daytime sitting.",
                "price": 22.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Date night babysitting", "Evening sitting", "Infant care", "Weekend care"]
            },
            {
                "name": "Nanny Services",
                "description": "Experienced full-time, part-time, and live-in nannies, mother helpers, and newborn care specialists.",
                "price": 28.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Full-time nanny", "Part-time nanny", "Live-in nanny", "Newborn care specialist"]
            },
            {
                "name": "Daycare",
                "description": "Licensed child care centers, home daycares, structured play, and early learning infant care.",
                "price": 35.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Licensed home daycare", "Group daycare", "Toddler care", "Infant daycare center"]
            },
            {
                "name": "Preschool Programs",
                "description": "Montessori and traditional preschool readiness, kindergarten preparation, and developmental learning.",
                "price": 40.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Montessori preschool", "Early literacy", "Kindergarten readiness", "Play-based learning"]
            },
            {
                "name": "After School Care",
                "description": "Supervised after-school care, homework assistance, school pickups, snacks, and enrichment activities.",
                "price": 25.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Homework help", "Pickup from school", "Supervised play", "Evening snacks"]
            },
            {
                "name": "School Break and Holiday Programs",
                "description": "Seasonal holiday camps, spring break care, summer day camps, and school vacation coverage.",
                "price": 40.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Summer day camp", "Spring break care", "Holiday drop-in programs"]
            }
        ]
    },
    {
        "name": "Adult and Senior Support",
        "description": "In-home daily living assistance, senior companionship, respite care, disability support, and care coordination.",
        "image": "categories/senior_support.jpg",
        "services": [
            {
                "name": "Personal Care and Daily Living Assistance",
                "description": "Assistance with bathing, dressing, grooming, mobility, meal preparation, and medication reminders.",
                "price": 32.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Bathing", "Dressing", "Meal assistance", "Mobility support", "Medication reminders"]
            },
            {
                "name": "Companionship",
                "description": "Friendly conversation, reading, accompanied walks, games, errands, and emotional engagement for seniors.",
                "price": 25.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Senior conversation", "Reading", "Board games", "Accompanied walks", "Social stimulation"]
            },
            {
                "name": "Respite Care",
                "description": "Temporary relief and coverage for primary family caregivers for evenings, weekends, or scheduled breaks.",
                "price": 35.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Family caregiver relief", "Overnight respite", "Weekend coverage"]
            },
            {
                "name": "Disability Support",
                "description": "Dedicated personal support for individuals with physical or cognitive disabilities to foster independence.",
                "price": 35.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Adaptive mobility aid", "Independent living support", "Transfer assistance"]
            },
            {
                "name": "Care Coordination",
                "description": "Navigating medical appointments, coordinating physical therapy, care plan setup, and family updates.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Medical appointment accompaniment", "Care plan management", "Family advocacy"]
            }
        ]
    },
    {
        "name": "Pet Services",
        "description": "Mobile dog grooming, pet sitting, daily dog walking, pet boarding, behavioral training, and vet transport.",
        "image": "categories/pet_services.jpg",
        "services": [
            {
                "name": "Pet Grooming",
                "description": "Mobile van dog and cat grooming, hydrobaths, breed haircuts, deshedding, and nail clipping.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile dog grooming", "Breed haircut", "Deshedding treatment", "Nail trim", "Cat grooming", "Flea bath"]
            },
            {
                "name": "Pet Sitting",
                "description": "In-home pet visits, cat feeding and playtime, puppy check-ins, and overnight stay pet sitting.",
                "price": 25.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["In-home pet sitting", "Cat drop-in visits", "Puppy care", "Overnight sitting"]
            },
            {
                "name": "Dog Walking",
                "description": "Private dog walks, energetic group pack walks, neighborhood trail excursions, and bathroom breaks.",
                "price": 20.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Private dog walks", "Group pack walks", "Trail adventures", "Exercise relief walks"]
            },
            {
                "name": "Pet Boarding and Daycare",
                "description": "Cage-free dog boarding, doggy daycare, supervised group play, and cozy overnight accommodations.",
                "price": 40.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Cage-free dog boarding", "Doggy daycare", "Supervised playgroup"]
            },
            {
                "name": "Pet Training",
                "description": "Puppy obedience, leash manners, aggressive behavior modification, and trick agility training.",
                "price": 70.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Puppy obedience", "Leash reactivity", "Agility training", "Behavioral correction"]
            },
            {
                "name": "Veterinary Care",
                "description": "Mobile vet visits, wellness checkups, core vaccinations, minor sickness exams, and health certificates.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile vet visits", "Routine pet checkups", "Vaccinations", "Wellness exams"]
            },
            {
                "name": "Pet Transportation",
                "description": "Safe pet taxi services for vet appointments, airport transport, and groomer drop-off and pickup.",
                "price": 30.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Vet trip rides", "Airport pet transport", "Groomer drop-off and pickup"]
            }
        ]
    },
    {
        "name": "Education and Lessons",
        "description": "Academic tutoring, SAT/ACT test prep, language learning, music lessons, art classes, and coding lessons.",
        "image": "categories/education.jpg",
        "services": [
            {
                "name": "Academic Tutoring",
                "description": "One-on-one K-12 and college tutoring in math, science, reading, writing, history, and chemistry.",
                "price": 45.00,
                "location": "ONLINE",
                "specialties": ["Reading", "Mathematics", "Science", "Specific grades", "Chemistry", "Physics", "History"]
            },
            {
                "name": "Test Preparation",
                "description": "Strategic coaching and practice exams for SAT, ACT, GRE, GMAT, AP tests, and state exams.",
                "price": 60.00,
                "location": "ONLINE",
                "specialties": ["SAT prep", "ACT prep", "GRE/GMAT", "AP exam prep", "State standardized tests"]
            },
            {
                "name": "Language Lessons",
                "description": "Conversational and fluent instruction in Spanish, French, Mandarin, English ESL, Arabic, and German.",
                "price": 40.00,
                "location": "ONLINE",
                "specialties": ["Spanish", "French", "Mandarin", "English ESL", "Arabic", "German", "Sign language"]
            },
            {
                "name": "Music Lessons",
                "description": "Private instrument instruction for piano, acoustic/electric guitar, violin, drums, and vocal training.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Piano lessons", "Guitar lessons", "Vocal coaching", "Violin", "Drum instruction"]
            },
            {
                "name": "Art Lessons",
                "description": "Instruction in drawing, oil painting, watercolor, ceramics, sculpting, and digital graphic illustration.",
                "price": 40.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Drawing", "Oil painting", "Watercolor", "Ceramics", "Digital illustration"]
            },
            {
                "name": "Dance Lessons",
                "description": "Private and group dance coaching in ballet, hip hop, salsa, ballroom, tap, and contemporary dance.",
                "price": 45.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Ballet", "Hip hop", "Salsa and bachata", "Contemporary dance", "Ballroom"]
            },
            {
                "name": "Computer and Digital Skills",
                "description": "Coding lessons for kids and adults (Python, JS), Excel mastery, senior digital literacy, and computer basics.",
                "price": 50.00,
                "location": "ONLINE",
                "specialties": ["Coding for kids", "Senior computer basics", "MS Office / Excel", "Graphic software"]
            },
            {
                "name": "Vocational and Trade Training",
                "description": "Foundational preparation and mentoring for electrical, carpentry, HVAC, plumbing, and culinary trades.",
                "price": 60.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Electrical apprentice prep", "Carpentry basics", "Culinary skills", "Auto mechanics"]
            }
        ]
    },
    {
        "name": "Healthcare and Mental Health",
        "description": "Primary healthcare, dentistry, optometry, in-home nursing, physical therapy, counseling, and medical testing.",
        "image": "categories/healthcare.jpg",
        "services": [
            {
                "name": "Primary Care",
                "description": "Routine medical wellness exams, chronic disease management, vitals monitoring, and health consults.",
                "price": 100.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Routine health checkup", "Blood pressure monitoring", "Wellness visits", "Preventative care"]
            },
            {
                "name": "Dental Care",
                "description": "Dental cleanings, oral examinations, teeth whitening, fillings, and mobile dental van care.",
                "price": 90.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Dental exams", "Teeth cleaning", "Fluoride treatment", "Cavity fillings", "Teeth whitening"]
            },
            {
                "name": "Eye Care",
                "description": "Comprehensive vision testing, eyeglasses fittings, contact lens exams, and glaucoma screenings.",
                "price": 75.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Vision exams", "Prescription glasses", "Contact lens fitting", "Glaucoma screening"]
            },
            {
                "name": "Home Nursing",
                "description": "Skilled in-home nursing, post-op wound care, IV infusions, injections, and medication management.",
                "price": 65.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wound care", "Post-operative care", "IV infusion", "Medication administration"]
            },
            {
                "name": "Physical Therapy",
                "description": "Musculoskeletal rehabilitation, joint mobility, sports recovery, and post-surgery physical therapy.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Post-surgery rehab", "Sports injury recovery", "Joint mobility", "Chronic back pain"]
            },
            {
                "name": "Occupational Therapy",
                "description": "Rehabilitation to restore daily living abilities, fine motor skills, and stroke recovery support.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Hand therapy", "Activities of daily living rehab", "Stroke recovery"]
            },
            {
                "name": "Speech Therapy",
                "description": "Diagnosis and therapy for speech delays, articulation, fluency, swallowing disorders, and stuttering.",
                "price": 80.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Articulation therapy", "Language development", "Swallowing therapy", "Stuttering"]
            },
            {
                "name": "Counseling and Psychotherapy",
                "description": "Licensed individual, couples, and family therapy for anxiety, depression, grief, and life transitions.",
                "price": 90.00,
                "location": "ONLINE",
                "specialties": ["Cognitive behavioral therapy", "Anxiety and depression", "Couples counseling", "Family therapy"]
            },
            {
                "name": "Psychiatry",
                "description": "Psychiatric medical evaluations, psychiatric consultations, and safe prescription medication management.",
                "price": 140.00,
                "location": "ONLINE",
                "specialties": ["Psychiatric evaluations", "Medication management", "Mental wellness consultations"]
            },
            {
                "name": "Nutrition and Dietitian Services",
                "description": "Personalized nutritional plans, medical diet consulting, weight management, and meal guidance.",
                "price": 70.00,
                "location": "ONLINE",
                "specialties": ["Weight management", "Diabetic meal plans", "Sports nutrition", "Gut health diets"]
            },
            {
                "name": "Medical Testing and Imaging",
                "description": "In-home blood draws (phlebotomy), mobile X-rays, ultrasound imaging, and diagnostic test panels.",
                "price": 95.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile blood draws", "X-ray imaging", "Ultrasound scans", "Rapid diagnostic tests"]
            },
            {
                "name": "Specialty Medical Care",
                "description": "Specialist consultations across dermatology, podiatry, cardiology, and allergy care.",
                "price": 125.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Dermatology consultations", "Podiatry foot care", "Cardiology checkups", "Allergy care"]
            }
        ]
    },
    {
        "name": "Food and Catering",
        "description": "Catering for events, in-home personal chefs, custom bakeries, meal prep delivery, and dessert bars.",
        "image": "categories/food_catering.jpg",
        "services": [
            {
                "name": "Catering",
                "description": "Full-service wedding and corporate catering, hot buffet stations, box lunches, and BBQ catering.",
                "price": 30.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Corporate buffet catering", "Wedding catering", "Box lunches", "Barbecue catering", "Plated dinners"]
            },
            {
                "name": "Personal Chef Services",
                "description": "Private in-home chef for intimate dinner parties, romantic dates, custom menus, and culinary events.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["In-home private chef", "Multi-course dinner parties", "Romantic dinners", "Dietary custom cooking"]
            },
            {
                "name": "Meal Preparation",
                "description": "Weekly customized healthy meal prep, portion-controlled containers, keto, vegan, and family meals.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Weekly meal prep", "Keto and low carb prep", "Vegan meal prep", "Family dinners"]
            },
            {
                "name": "Bakery Services",
                "description": "Custom birthday cakes, wedding cakes, fresh artisan breads, donuts, pastries, and gluten-free baking.",
                "price": 40.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Donuts", "Birthday cakes", "Wedding cakes", "Pastries", "Artisan bread", "Gluten-free baking"]
            },
            {
                "name": "Desserts and Frozen Treats",
                "description": "Ice cream carts, cupcake towers, gourmet cookies, dessert tables, churro stations, and crepe bars.",
                "price": 35.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Ice cream carts", "Custom cookies", "Cupcakes", "Dessert tables", "Churros and crepes"]
            },
            {
                "name": "Coffee and Beverage Services",
                "description": "Mobile espresso bars, barista service, craft mocktail stations, fresh smoothie carts, and boba tea.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile espresso bar", "Smoothies and juice bar", "Boba tea bar", "Craft mocktails"]
            },
            {
                "name": "Restaurants and Prepared Meals",
                "description": "Group family feast packages, artisan takeout catering platters, and hot prepared meals.",
                "price": 25.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Takeout catering", "Family feast platters", "Gourmet ready-to-eat dishes"]
            }
        ]
    },
    {
        "name": "Events and Entertainment",
        "description": "Event coordination, party rentals, floral decor, live DJs, musicians, performers, and professional bartending.",
        "image": "categories/events_entertainment.jpg",
        "services": [
            {
                "name": "Event Planning",
                "description": "Full-service wedding planning, corporate gala coordination, birthday party management, and design.",
                "price": 80.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wedding planning", "Corporate galas", "Birthday party coordination", "Full-service event design"]
            },
            {
                "name": "Event Decoration",
                "description": "Balloon arches, floral centerpieces, photo backdrops, mood lighting, and aesthetic table styling.",
                "price": 60.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Balloon arches", "Floral arrangements", "Backdrops and photo booths", "Table centerpieces"]
            },
            {
                "name": "Event Rentals",
                "description": "Tables, chairs, party tents, dance floors, linens, staging, and audiovisual rental equipment.",
                "price": 50.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Tables and chairs", "Tent rentals", "Linen and tableware", "Dance floors", "Audio/Visual gear"]
            },
            {
                "name": "Venue Rental",
                "description": "Private event venues, scenic gardens, banquet halls, rustic barns, and modern urban loft spaces.",
                "price": 150.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Banquet halls", "Outdoor garden venues", "Loft event spaces", "Meeting rooms"]
            },
            {
                "name": "DJs",
                "description": "Live event DJs, wedding sound systems, MC hosting, dance party mixes, and intelligent lighting.",
                "price": 90.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wedding DJ", "Corporate event DJ", "Club and party DJ", "Karaoke hosting"]
            },
            {
                "name": "Live Performers",
                "description": "Live bands, acoustic soloists, string quartets, magicians, comedians, and specialty musicians.",
                "price": 100.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Live bands", "Acoustic soloists", "String quartets", "Magicians", "Vocalists"]
            },
            {
                "name": "Party Entertainment",
                "description": "Face painters, balloon artists, costumed character visits, clowns, and inflatable bounce houses.",
                "price": 55.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Face painting", "Balloon twisting", "Character visits", "Clowns", "Inflatable bounce houses"]
            },
            {
                "name": "Bartending",
                "description": "Certified event mixologists, mobile pop-up bar setups, signature cocktail creation, and beverage pouring.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile bar service", "Craft cocktail mixologists", "Beer and wine pouring", "Event bar setup"]
            }
        ]
    },
    {
        "name": "Photography and Creative Services",
        "description": "Professional photography, videography, photo retouching, graphic design, copywriting, and printing.",
        "image": "categories/photography.jpg",
        "services": [
            {
                "name": "Photography",
                "description": "Wedding photography, family portraits, newborn sessions, corporate headshots, and real estate photos.",
                "price": 95.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Weddings", "Family portraits", "Newborns", "Events", "Product photography", "Headshots", "Real estate"]
            },
            {
                "name": "Videography",
                "description": "Cinematic wedding films, corporate promos, drone aerial videography, social media reels, and event videos.",
                "price": 120.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Wedding films", "Commercial video", "Drone aerial footage", "Social media reels", "Event recaps"]
            },
            {
                "name": "Photo and Video Editing",
                "description": "Professional photo retouching, color grading, sound design, video cutting, and album creation.",
                "price": 50.00,
                "location": "ONLINE",
                "specialties": ["Color grading", "Retouching", "Video post-production", "Album design", "Sound editing"]
            },
            {
                "name": "Graphic Design and Branding",
                "description": "Logo design, comprehensive brand guidelines, business cards, flyers, and digital promotional graphics.",
                "price": 55.00,
                "location": "ONLINE",
                "specialties": ["Logo design", "Brand identity", "Business cards", "Brochures and flyers", "UI design"]
            },
            {
                "name": "Writing and Editing",
                "description": "Professional copywriting, website content, resume writing, blog posts, proofreading, and ghostwriting.",
                "price": 45.00,
                "location": "ONLINE",
                "specialties": ["Copywriting", "Technical writing", "Proofreading", "Blog articles", "Ghostwriting"]
            },
            {
                "name": "Printing and Publishing",
                "description": "Large format poster printing, custom vinyl banners, branded apparel, business brochures, and book binding.",
                "price": 40.00,
                "location": "PROVIDER_LOCATION",
                "specialties": ["Banner printing", "Custom apparel printing", "Large format prints", "Book binding"]
            }
        ]
    },
    {
        "name": "Business and Professional Services",
        "description": "Accounting, tax preparation, legal contracts, notary public, consulting, real estate, and digital marketing.",
        "image": "categories/business_services.jpg",
        "services": [
            {
                "name": "Accounting and Bookkeeping",
                "description": "Monthly business bookkeeping, QuickBooks management, financial reporting, and payroll processing.",
                "price": 65.00,
                "location": "ONLINE",
                "specialties": ["Quickbooks setup", "Monthly bookkeeping", "Financial statement preparation", "Payroll"]
            },
            {
                "name": "Tax Preparation",
                "description": "Individual tax return filing, corporate business taxes, tax deduction planning, and IRS advisory.",
                "price": 85.00,
                "location": "ONLINE",
                "specialties": ["Individual tax returns", "Business corporate taxes", "Tax planning", "IRS resolution"]
            },
            {
                "name": "Legal Services",
                "description": "Business contract drafting, LLC formation, estate planning, wills, and trademark registration.",
                "price": 150.00,
                "location": "ONLINE",
                "specialties": ["Contract review", "Business formation", "Estate planning and wills", "Trademark filing"]
            },
            {
                "name": "Notary Services",
                "description": "Mobile notary public, loan signing agent, document witnessing, apostille services, and power of attorney.",
                "price": 45.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mobile notary public", "Loan signing agent", "Document witnessing", "Apostille assistance"]
            },
            {
                "name": "Business Consulting",
                "description": "Strategic business growth, startup launch advisory, operations efficiency, and financial modeling.",
                "price": 100.00,
                "location": "ONLINE",
                "specialties": ["Strategic planning", "Startup advisory", "Operations optimization", "Franchise consulting"]
            },
            {
                "name": "Real Estate Services",
                "description": "Home buying/selling representation, property management, home staging, and rental leasing.",
                "price": 100.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Buyer representation", "Home staging", "Property management", "Rental leasing"]
            },
            {
                "name": "Resume and Career Services",
                "description": "Executive resume writing, LinkedIn profile optimization, interview preparation, and career coaching.",
                "price": 60.00,
                "location": "ONLINE",
                "specialties": ["Resume writing", "LinkedIn profile optimization", "Interview coaching", "Career coaching"]
            },
            {
                "name": "Translation and Interpretation",
                "description": "Certified document translation, legal transcription, and real-time live language interpretation.",
                "price": 50.00,
                "location": "ONLINE",
                "specialties": ["Document translation", "In-person court interpretation", "Medical translation"]
            },
            {
                "name": "Virtual Assistance",
                "description": "Remote administrative assistance, calendar scheduling, inbox management, and data entry.",
                "price": 25.00,
                "location": "ONLINE",
                "specialties": ["Email and calendar management", "Data entry", "Customer service support", "Research tasks"]
            },
            {
                "name": "Marketing Services",
                "description": "Social media management, SEO optimization, Google and Facebook paid ads, and email campaigns.",
                "price": 75.00,
                "location": "ONLINE",
                "specialties": ["Social media", "SEO", "Paid advertising", "Email marketing", "Influencer outreach", "Content creation"]
            }
        ]
    },
    {
        "name": "Technology Services",
        "description": "Device repair, IT technical support, home Wi-Fi networks, TV mounting, smart home setup, and web/app development.",
        "image": "categories/technology.jpg",
        "services": [
            {
                "name": "Computer and Device Repair",
                "description": "PC and Mac repair, laptop screen replacement, virus removal, hardware upgrades, and data recovery.",
                "price": 70.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Laptop screen repair", "PC troubleshooting", "Virus and malware removal", "Data recovery"]
            },
            {
                "name": "Technical Support",
                "description": "Remote and on-site IT help, software installations, printer troubleshooting, and operating system updates.",
                "price": 55.00,
                "location": "ONLINE",
                "specialties": ["Printer setup", "Remote desktop support", "Software installation", "OS upgrades"]
            },
            {
                "name": "Home Networking and Wi Fi",
                "description": "Mesh Wi-Fi system setup, router optimization, Ethernet cabling, and dead-zone elimination.",
                "price": 75.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Mesh Wi-Fi setup", "Ethernet cabling", "Router configuration", "Dead zone elimination"]
            },
            {
                "name": "TV and Home Theater Installation",
                "description": "Wall TV mounting, concealed wiring, soundbar installation, surround sound, and home cinema setup.",
                "price": 85.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["TV wall mounting", "Surround sound setup", "Projector installation", "Wire concealment"]
            },
            {
                "name": "Smart Home and Security Installation",
                "description": "Video doorbell installation, security camera setup, smart thermostats, smart locks, and home automation.",
                "price": 90.00,
                "location": "CUSTOMER_LOCATION",
                "specialties": ["Ring doorbell setup", "Smart thermostat", "Security camera installation", "Smart locks"]
            },
            {
                "name": "Website Development",
                "description": "Custom responsive website design, WordPress sites, Shopify e-commerce, and web application development.",
                "price": 80.00,
                "location": "ONLINE",
                "specialties": ["WordPress websites", "E-commerce Shopify stores", "Custom web applications", "Responsive design"]
            },
            {
                "name": "App and Software Development",
                "description": "iOS and Android mobile app development, backend APIs, database architecture, and custom software systems.",
                "price": 100.00,
                "location": "ONLINE",
                "specialties": ["iOS and Android mobile apps", "Backend API development", "Database design", "Cloud deployment"]
            },
            {
                "name": "Automation and Systems Integration",
                "description": "Zapier workflow automation, CRM integrations, API webhook connections, and AI chatbot setups.",
                "price": 90.00,
                "location": "ONLINE",
                "specialties": ["Zapier and webhook automation", "CRM integration", "API workflows", "AI chatbot integration"]
            }
        ]
    }
]

def generate():
    go_code = '''package main

import (
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryData struct {
	Name        string
	Description string
	Image       string
	Services    []ServiceData
}

type ServiceData struct {
	Name            string
	Description     string
	BasePrice       float64
	ServiceLocation string
	Specialties     []string
}

var categoriesSeed = []CategoryData{
'''

    for cat in categories_data:
        go_code += '\t{\n'
        go_code += f'\t\tName:        "{cat["name"]}",\n'
        go_code += f'\t\tDescription: "{cat["description"]}",\n'
        go_code += f'\t\tImage:       "{cat["image"]}",\n'
        go_code += '\t\tServices: []ServiceData{\n'
        for s in cat['services']:
            specs_str = ', '.join(f'"{x}"' for x in s['specialties'])
            go_code += f'\t\t\t{{"{s["name"]}", "{s["description"]}", {s["price"]:.2f}, "{s["location"]}", []string{{{specs_str}}}}},\n'
        go_code += '\t\t},\n\t},\n'

    go_code += '''}

func main() {
	log.Println("🌱 ====================================================================")
	log.Println("🌱 Neighbor Service - Syncing Official Categories & Catalog Services")
	log.Println("🌱 Source of Truth: NS Category Listing Documentation")
	log.Println("🌱 ====================================================================")

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// AutoMigrate tables to ensure schema parity with Specialties and Location attributes
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	log.Println("🧹 Cleaning up previous legacy seeded categories & catalog services...")
	// Clear previous mappings and catalog services to ensure pristine alignment
	db.Exec("DELETE FROM accounts_profile_catalog_services")
	db.Exec("UPDATE services_servicerequest SET catalog_service_id = NULL")
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.CatalogService{}).Error; err != nil {
		log.Printf("⚠️ Warning deleting catalog services: %v", err)
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Category{}).Error; err != nil {
		log.Printf("⚠️ Warning deleting categories: %v", err)
	}

	totalCategories := 0
	totalServices := 0

	for _, catData := range categoriesSeed {
		categoryID := uuid.New()
		cat := entity.Category{
			ID:          categoryID,
			Name:        catData.Name,
			Description: catData.Description,
			Image:       catData.Image,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}

		if err := db.Create(&cat).Error; err != nil {
			log.Fatalf("❌ Failed to create category %s: %v", cat.Name, err)
		}
		totalCategories++

		for _, srvData := range catData.Services {
			price := srvData.BasePrice
			catalogService := entity.CatalogService{
				ID:                     uuid.New(),
				CategoryID:             categoryID,
				Name:                   srvData.Name,
				Description:            srvData.Description,
				BasePrice:              &price,
				Specialties:            entity.JSONSlice(srvData.Specialties),
				DefaultServiceLocation: srvData.ServiceLocation,
				CreatedAt:              time.Now().UTC(),
				UpdatedAt:              time.Now().UTC(),
			}

			if err := db.Create(&catalogService).Error; err != nil {
				log.Fatalf("❌ Failed to create catalog service %s: %v", srvData.Name, err)
			}
			totalServices++
		}
	}

	log.Printf("✅ Successfully seeded %d Parent Categories and %d Child Catalog Services with Searchable Specialties!", totalCategories, totalServices)
}
'''

    with open("backend-go/cmd/seed_services/main.go", "w") as f:
        f.write(go_code)

    total_srvs = sum(len(c["services"]) for c in categories_data)
    print(f"Successfully generated main.go with {len(categories_data)} categories and {total_srvs} services.")

if __name__ == "__main__":
    generate()
