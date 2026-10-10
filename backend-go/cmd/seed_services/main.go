package main

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
	{
		Name:        "Home Repairs and Improvements",
		Description: "Professional residential and commercial repairs, plumbing, electrical, HVAC, carpentry, painting, roofing, and remodeling services.",
		Image:       "categories/home_repairs.jpg",
		Services: []ServiceData{
			{"Handyman Services", "General home repairs, fixture installations, drywall patching, furniture assembly, and small maintenance projects.", 50.00, "CUSTOMER_LOCATION", []string{"Drywall repair", "Furniture assembly", "TV mounting", "Fixture installation", "Door repair", "Minor home repairs"}},
			{"Plumbing", "Certified plumbing repairs, leak detection, drain clearing, water heater installation, and pipe maintenance.", 85.00, "CUSTOMER_LOCATION", []string{"Leak repair", "Drain cleaning", "Water heaters", "Fixture installation", "Pipe repair", "Toilet repair"}},
			{"Electrical Services", "Licensed electrical wiring, outlet repair, panel upgrades, lighting installations, and EV charger setup.", 95.00, "CUSTOMER_LOCATION", []string{"Light fixture installation", "Ceiling fans", "Outlet and switch repair", "Panel upgrades", "Wiring", "EV charger installation"}},
			{"Heating and Air Conditioning", "HVAC system inspection, heating repair, AC maintenance, duct cleaning, and heat pump installations.", 110.00, "CUSTOMER_LOCATION", []string{"AC repair", "Furnace maintenance", "HVAC installation", "Duct cleaning", "Thermostat installation", "Heat pump service"}},
			{"Appliance Repair", "Diagnosis and repair for refrigerators, washers, dryers, dishwashers, ovens, and household appliances.", 75.00, "CUSTOMER_LOCATION", []string{"Refrigerator repair", "Washer and dryer repair", "Dishwasher repair", "Oven and stove repair", "Microwave repair"}},
			{"Carpentry", "Custom cabinetry, trim carpentry, baseboards, crown molding, framing, and wood structural repairs.", 70.00, "CUSTOMER_LOCATION", []string{"Custom cabinetry", "Trim and molding", "Framing", "Deck repair", "Wood rot repair", "Shelving"}},
			{"Painting", "Interior and exterior house painting, cabinet refinishing, deck staining, and wallpaper removal.", 60.00, "CUSTOMER_LOCATION", []string{"Interior painting", "Exterior painting", "Cabinet painting", "Deck staining", "Drywall patching", "Wallpaper removal"}},
			{"Flooring", "Hardwood floor installation, tile and grout work, vinyl plank, laminate, and carpet laying.", 65.00, "CUSTOMER_LOCATION", []string{"Hardwood installation", "Tile and grout", "Laminate flooring", "Carpet installation", "Floor refinishing", "Vinyl plank"}},
			{"Roofing", "Roof leak repair, shingle replacement, flashing maintenance, inspections, and full roof repairs.", 120.00, "CUSTOMER_LOCATION", []string{"Roof leak repair", "Shingle replacement", "Roof inspections", "Flashing repair", "Gutter integration", "Flat roofing"}},
			{"Remodeling", "Full kitchen and bathroom remodels, basement finishing, room conversions, and home renovations.", 150.00, "CUSTOMER_LOCATION", []string{"Kitchen remodeling", "Bathroom remodeling", "Basement finishing", "Room additions", "Attic conversion"}},
			{"Doors and Windows", "Window installation, glass replacement, exterior and interior door fitting, sliding doors, and screens.", 80.00, "CUSTOMER_LOCATION", []string{"Window replacement", "Exterior door installation", "Sliding door repair", "Storm doors", "Screen replacement"}},
			{"Locksmith Services", "Emergency lockouts, rekeying locks, smart lock installations, deadbolts, and security hardware.", 75.00, "CUSTOMER_LOCATION", []string{"Emergency lockout", "Rekeying", "Smart lock installation", "Deadbolt repair", "Master key systems"}},
		},
	},
	{
		Name:        "Cleaning and Household Services",
		Description: "Residential home cleaning, deep cleaning, commercial janitorial, pressure washing, pest control, and junk removal.",
		Image:       "categories/cleaning.jpg",
		Services: []ServiceData{
			{"House Cleaning", "Recurring maid service, deep cleaning, move-in/move-out cleans, and eco-friendly home sanitization.", 45.00, "CUSTOMER_LOCATION", []string{"Standard recurring cleaning", "Deep cleaning", "Move-in/move-out cleaning", "Eco-friendly cleaning", "Post-renovation"}},
			{"Commercial Cleaning", "Office cleaning, retail space sanitization, janitorial maintenance, and commercial facility care.", 60.00, "CUSTOMER_LOCATION", []string{"Office cleaning", "Janitorial services", "Retail cleaning", "Medical office sanitization"}},
			{"Carpet and Upholstery Cleaning", "Deep steam carpet cleaning, sofa and couch shampooing, stain removal, and pet odor treatments.", 70.00, "CUSTOMER_LOCATION", []string{"Steam cleaning", "Stain removal", "Pet odor removal", "Sofa and couch cleaning", "Rug cleaning"}},
			{"Window Cleaning", "Streak-free interior and exterior window washing, screen scrubbing, and skylight cleaning.", 50.00, "CUSTOMER_LOCATION", []string{"Interior window washing", "Exterior window washing", "Screen cleaning", "Skylight cleaning"}},
			{"Pressure Washing", "High-pressure washing for driveways, sidewalks, home siding, patios, decks, and fences.", 80.00, "CUSTOMER_LOCATION", []string{"Driveway cleaning", "Siding power wash", "Deck and patio wash", "Roof soft wash", "Fence wash"}},
			{"Laundry and Dry Cleaning", "Wash, dry, fold services, dry cleaning pickup and delivery, garment pressing, and ironing.", 30.00, "CUSTOMER_LOCATION", []string{"Wash and fold", "Dry cleaning pickup", "Ironing", "Garment care", "Bedding and linen"}},
			{"Home Organization", "Professional organizing for closets, pantries, garages, kitchens, and decluttering assistance.", 55.00, "CUSTOMER_LOCATION", []string{"Closet organization", "Garage cleanout", "Kitchen pantry organizing", "Decluttering assistance"}},
			{"Pest Control", "Extermination and preventative treatments for termites, rodents, bed bugs, ants, and roaches.", 90.00, "CUSTOMER_LOCATION", []string{"Termite treatment", "Rodent control", "Bed bug extermination", "Ant and roach treatment", "Mosquito fogging"}},
			{"Junk Removal", "Eco-friendly hauling of old furniture, appliances, yard waste, construction debris, and estate cleanouts.", 75.00, "CUSTOMER_LOCATION", []string{"Furniture haul away", "Appliance disposal", "Estate cleanout", "Yard waste removal", "Construction debris"}},
		},
	},
	{
		Name:        "Lawn and Outdoor Services",
		Description: "Lawn mowing, landscaping design, tree trimming, irrigation systems, fences, patios, and pool care.",
		Image:       "categories/lawn_outdoor.jpg",
		Services: []ServiceData{
			{"Lawn Care", "Lawn mowing, edging, weed trimming, fertilization, aeration, overseeding, and seasonal yard cleanup.", 40.00, "CUSTOMER_LOCATION", []string{"Lawn mowing", "Edging", "Fertilization", "Aeration and overseeding", "Weed control"}},
			{"Landscaping", "Landscape design, plant installation, mulching, sod laying, garden bed edging, and rock features.", 65.00, "CUSTOMER_LOCATION", []string{"Landscape design", "Mulching", "Sod installation", "Plant and shrub planting", "Retaining walls"}},
			{"Tree Care", "Tree trimming, branch pruning, tree removal, stump grinding, and certified arborist consultations.", 120.00, "CUSTOMER_LOCATION", []string{"Tree trimming", "Tree removal", "Stump grinding", "Arborist consulting", "Emergency branch removal"}},
			{"Irrigation and Sprinklers", "Sprinkler system installation, zone repairs, head replacements, drip lines, and winterization.", 75.00, "CUSTOMER_LOCATION", []string{"Sprinkler system repair", "Drip irrigation", "Winterization", "Backflow testing", "Timer programming"}},
			{"Fence and Gate Services", "Wood, vinyl, and chain link fence construction, post replacements, gate repairs, and automation.", 85.00, "CUSTOMER_LOCATION", []string{"Wood fence installation", "Chain link", "Vinyl fencing", "Gate repair", "Automatic gate openers"}},
			{"Decks and Patios", "Custom deck building, paver patios, concrete stamping, porch repairs, and pergola construction.", 130.00, "CUSTOMER_LOCATION", []string{"Paver patio construction", "Deck building", "Concrete pouring", "Pergolas and gazebos"}},
			{"Pool Maintenance", "Weekly pool cleaning, chemical balancing, filter backwashing, pump repair, and equipment service.", 80.00, "CUSTOMER_LOCATION", []string{"Chemical balancing", "Pool opening/closing", "Filter cleaning", "Leak detection", "Tile scrubbing"}},
			{"Gutter Cleaning", "Debris removal from gutters and downspouts, gutter guard installation, and leak patching.", 60.00, "CUSTOMER_LOCATION", []string{"Gutter debris removal", "Gutter guard installation", "Downspout flushing", "Minor gutter repair"}},
		},
	},
	{
		Name:        "Moving and Storage",
		Description: "Residential moving, commercial office moves, packing/unpacking, furniture assembly, and storage solutions.",
		Image:       "categories/moving_storage.jpg",
		Services: []ServiceData{
			{"Residential Moving", "Local and long-distance home moving, apartment relocations, and professional furniture handling.", 95.00, "CUSTOMER_LOCATION", []string{"Local home moving", "Long distance moving", "Apartment moving", "Senior relocation"}},
			{"Commercial Moving", "Office relocations, cubicle moves, corporate equipment transfer, and business transitions.", 120.00, "CUSTOMER_LOCATION", []string{"Office relocation", "Business equipment move", "Retail store moving"}},
			{"Packing and Unpacking", "Full-service packing with boxes, bubble wrap, fragile crating, and organized unpacking support.", 50.00, "CUSTOMER_LOCATION", []string{"Full packing service", "Fragile item crating", "Box labeling", "Unpacking assistance"}},
			{"Furniture Assembly", "Assembly of IKEA flat-pack furniture, bed frames, dressers, desks, and home gym equipment.", 45.00, "CUSTOMER_LOCATION", []string{"IKEA furniture assembly", "Bed frames", "Desks and tables", "Exercise equipment assembly"}},
			{"Loading and Unloading", "Heavy lifting labor for loading or unloading moving trucks, Pods, trailers, and storage units.", 45.00, "CUSTOMER_LOCATION", []string{"Truck loading labor", "Storage container loading", "Heavy lifting labor"}},
			{"Storage Services", "Short and long term secure storage units, climate control, vehicle storage, and portable pods.", 60.00, "PROVIDER_LOCATION", []string{"Secure unit storage", "Climate controlled storage", "Vehicle storage", "Mobile storage delivery"}},
		},
	},
	{
		Name:        "Automotive Services",
		Description: "Mobile mechanics, car detailing, mobile tire service, auto glass, battery replacement, and towing.",
		Image:       "categories/automotive.jpg",
		Services: []ServiceData{
			{"Auto Repair and Maintenance", "On-site brake pad replacement, oil changes, engine diagnostic scans, starters, and alternators.", 80.00, "CUSTOMER_LOCATION", []string{"Brake service", "Oil change", "Engine diagnostics", "Transmission repair", "Tune-ups", "Alternators"}},
			{"Car Washing and Detailing", "Mobile hand car wash, interior shampooing, exterior buffing, clay bar, and ceramic coating.", 55.00, "CUSTOMER_LOCATION", []string{"Mobile car wash", "Interior deep detail", "Paint correction", "Ceramic coating", "Leather conditioning"}},
			{"Tire Services", "Mobile tire mounting, balancing, flat tire puncture repairs, and tire rotation at your location.", 50.00, "CUSTOMER_LOCATION", []string{"Tire mounting", "Balancing", "Flat tire repair", "Tire rotation", "Mobile tire change"}},
			{"Auto Glass Services", "Windshield chip repairs, cracked glass replacement, side window and rearview mirror replacements.", 75.00, "CUSTOMER_LOCATION", []string{"Windshield chip repair", "Windshield replacement", "Side glass", "Mirror replacement"}},
			{"Battery Services", "On-demand battery jump starts, battery diagnostic health tests, and new battery installations.", 45.00, "CUSTOMER_LOCATION", []string{"Battery jump start", "Battery testing", "On-site battery replacement", "Terminal cleaning"}},
			{"Roadside Assistance", "Emergency roadside help including lockout rescue, emergency fuel delivery, and flat tire swaps.", 60.00, "CUSTOMER_LOCATION", []string{"Lockout service", "Fuel delivery", "Winch out", "Flat tire change"}},
			{"Towing", "Emergency flatbed and wheel-lift towing for cars, SUVs, motorcycles, and disabled vehicles.", 95.00, "CUSTOMER_LOCATION", []string{"Flatbed towing", "Wheel lift towing", "Long distance towing", "Accident recovery"}},
			{"Vehicle Inspections", "Comprehensive pre-purchase mechanical inspections, safety checks, and computer diagnostics.", 70.00, "CUSTOMER_LOCATION", []string{"Pre-purchase inspections", "Emissions check", "Safety inspection", "Diagnostic scanning"}},
		},
	},
	{
		Name:        "Transportation and Delivery",
		Description: "Passenger transit, private chauffeurs, airport shuttles, couriers, freight delivery, and errand runs.",
		Image:       "categories/transportation.jpg",
		Services: []ServiceData{
			{"Passenger Transportation", "Reliable point-to-point passenger rides, private transportation, and non-emergency transit.", 35.00, "CUSTOMER_LOCATION", []string{"Local rides", "Point-to-point transit", "Group shuttle", "Non-emergency medical transit"}},
			{"Chauffeur Services", "Executive black car services, luxury sedan travel, dedicated personal drivers, and VIP transit.", 75.00, "CUSTOMER_LOCATION", []string{"Executive black car", "Luxury sedan", "Hourly private driver", "Special event transport"}},
			{"Airport Transfers", "Scheduled airport departures and pickups, luggage assistance, and flight status tracking.", 55.00, "CUSTOMER_LOCATION", []string{"Airport drop-off", "Airport pickup", "Flight tracking arrival", "Luggage assistance"}},
			{"Courier Services", "Same-day legal document deliveries, medical parcel courier, and urgent package distribution.", 30.00, "CUSTOMER_LOCATION", []string{"Same-day legal documents", "Medical sample courier", "Package delivery", "Express parcel"}},
			{"Grocery and Retail Delivery", "Personal supermarket shopping, grocery delivery, and store merchandise pickup to your door.", 25.00, "CUSTOMER_LOCATION", []string{"Supermarket grocery run", "Pharmacy medicine pickup", "Retail merchandise delivery"}},
			{"Freight and Large Item Delivery", "Cargo van transport, appliance delivery, pallet moving, and oversized goods transportation.", 85.00, "CUSTOMER_LOCATION", []string{"Pallet transport", "Large appliance delivery", "Furniture transport", "Cargo van transport"}},
			{"Errand Services", "Personal assistant errands, post office drops, prescription pickups, and general to-do tasks.", 25.00, "CUSTOMER_LOCATION", []string{"Personal errands", "Post office runs", "Shopping assistance", "Dry cleaning pickup"}},
		},
	},
	{
		Name:        "Beauty and Personal Style",
		Description: "Hair styling, barbering, nail care, makeup artistry, facials, waxing, lash extensions, and personal styling.",
		Image:       "categories/beauty_grooming.jpg",
		Services: []ServiceData{
			{"Hair Services", "Haircuts, blowouts, hair coloring, balayage, braiding, natural hair care, locs, and extensions.", 55.00, "CUSTOMER_LOCATION", []string{"Braiding", "Natural hair", "Coloring", "Extensions", "Locs", "Blowouts", "Haircuts", "Balayage"}},
			{"Barbering", "Classic fades, men haircuts, beard sculpting, line-ups, and luxury hot towel straight-razor shaves.", 35.00, "CUSTOMER_LOCATION", []string{"Men haircuts", "Fades", "Beard sculpting", "Hot towel shave", "Edge ups", "Lineups"}},
			{"Nail Care", "Manicures, spa pedicures, acrylic full sets, gel polish, dip powder, and custom nail art.", 40.00, "CUSTOMER_LOCATION", []string{"Manicures", "Pedicures", "Gel nails", "Acrylic full sets", "Dip powder", "Nail art"}},
			{"Makeup", "Bridal makeup, glam makeup for parties, photoshoot beauty, and special occasion styling.", 65.00, "CUSTOMER_LOCATION", []string{"Bridal makeup", "Glam makeup", "Photoshoot makeup", "Special event makeup", "SFX makeup"}},
			{"Skin Care and Facials", "Hydrating facials, deep pore cleansing, microdermabrasion, anti-aging therapies, and acne care.", 60.00, "CUSTOMER_LOCATION", []string{"Hydrating facials", "Chemical peels", "Microdermabrasion", "Anti-aging treatments", "Acne care"}},
			{"Waxing and Hair Removal", "Full body waxing, Brazilian wax, precision eyebrow threading, sugaring, and laser hair removal.", 45.00, "CUSTOMER_LOCATION", []string{"Brazilian waxing", "Eyebrow threading", "Body waxing", "Laser hair removal", "Sugaring"}},
			{"Lash and Brow Services", "Classic and volume lash extensions, lash lifts, brow lamination, microblading, and brow tinting.", 55.00, "CUSTOMER_LOCATION", []string{"Lash extensions", "Lash lift and tint", "Brow lamination", "Microblading", "Brow shaping"}},
			{"Personal Styling", "Wardrobe consulting, color palette analysis, personal shopping, closet audits, and event styling.", 70.00, "CUSTOMER_LOCATION", []string{"Wardrobe styling", "Color analysis", "Personal shopping", "Closet audit", "Event outfit curation"}},
			{"Tailoring and Dressmaking", "Clothing alterations, hemming pants, custom suit tailoring, dress adjustments, and zipper repair.", 35.00, "PROVIDER_LOCATION", []string{"Alterations", "Hemming", "Custom suits", "Bridal gown alterations", "Zipper replacement"}},
		},
	},
	{
		Name:        "Fitness and Wellness",
		Description: "Personal fitness training, group classes, yoga, sports coaching, massage therapy, and wellness coaching.",
		Image:       "categories/fitness_wellness.jpg",
		Services: []ServiceData{
			{"Personal Training", "One-on-one fitness training, strength conditioning, fat loss programs, and home workout routines.", 60.00, "CUSTOMER_LOCATION", []string{"One-on-one training", "Weight loss", "Strength and conditioning", "Mobility training", "In-home fitness"}},
			{"Fitness Classes", "High-intensity interval training (HIIT), bootcamps, spin cycling, Zumba, and cardio fitness classes.", 25.00, "PROVIDER_LOCATION", []string{"HIIT classes", "Bootcamps", "Group cycling", "Zumba", "Barre classes"}},
			{"Yoga and Pilates", "Vinyasa yoga, Hatha yoga, Mat and Reformer Pilates, meditation, and guided breathwork sessions.", 50.00, "CUSTOMER_LOCATION", []string{"Vinyasa yoga", "Hatha yoga", "Mat Pilates", "Reformer Pilates", "Breathwork", "Meditation"}},
			{"Sports Coaching", "Private and youth coaching for tennis, swimming, basketball, soccer, golf, and athletic skills.", 55.00, "CUSTOMER_LOCATION", []string{"Tennis lessons", "Swimming coaching", "Golf instruction", "Soccer skills", "Basketball training"}},
			{"Massage Therapy", "In-home deep tissue massage, Swedish relaxation, sports massage, prenatal massage, and trigger point therapy.", 85.00, "CUSTOMER_LOCATION", []string{"Deep tissue massage", "Swedish massage", "Sports massage", "Prenatal massage", "Hot stone therapy"}},
			{"Wellness Coaching", "Lifestyle wellness, holistic health guidance, stress relief coaching, and sustainable habit formation.", 65.00, "ONLINE", []string{"Holistic lifestyle coaching", "Stress management", "Sleep coaching", "Habit formation"}},
		},
	},
	{
		Name:        "Childcare and Early Learning",
		Description: "Vetted babysitters, full-time nannies, licensed daycares, preschool programs, and after-school care.",
		Image:       "categories/childcare.jpg",
		Services: []ServiceData{
			{"Babysitting", "Trusted on-demand babysitting, date-night childcare, infant care, and weekend daytime sitting.", 22.00, "CUSTOMER_LOCATION", []string{"Date night babysitting", "Evening sitting", "Infant care", "Weekend care"}},
			{"Nanny Services", "Experienced full-time, part-time, and live-in nannies, mother helpers, and newborn care specialists.", 28.00, "CUSTOMER_LOCATION", []string{"Full-time nanny", "Part-time nanny", "Live-in nanny", "Newborn care specialist"}},
			{"Daycare", "Licensed child care centers, home daycares, structured play, and early learning infant care.", 35.00, "PROVIDER_LOCATION", []string{"Licensed home daycare", "Group daycare", "Toddler care", "Infant daycare center"}},
			{"Preschool Programs", "Montessori and traditional preschool readiness, kindergarten preparation, and developmental learning.", 40.00, "PROVIDER_LOCATION", []string{"Montessori preschool", "Early literacy", "Kindergarten readiness", "Play-based learning"}},
			{"After School Care", "Supervised after-school care, homework assistance, school pickups, snacks, and enrichment activities.", 25.00, "CUSTOMER_LOCATION", []string{"Homework help", "Pickup from school", "Supervised play", "Evening snacks"}},
			{"School Break and Holiday Programs", "Seasonal holiday camps, spring break care, summer day camps, and school vacation coverage.", 40.00, "PROVIDER_LOCATION", []string{"Summer day camp", "Spring break care", "Holiday drop-in programs"}},
		},
	},
	{
		Name:        "Adult and Senior Support",
		Description: "In-home daily living assistance, senior companionship, respite care, disability support, and care coordination.",
		Image:       "categories/senior_support.jpg",
		Services: []ServiceData{
			{"Personal Care and Daily Living Assistance", "Assistance with bathing, dressing, grooming, mobility, meal preparation, and medication reminders.", 32.00, "CUSTOMER_LOCATION", []string{"Bathing", "Dressing", "Meal assistance", "Mobility support", "Medication reminders"}},
			{"Companionship", "Friendly conversation, reading, accompanied walks, games, errands, and emotional engagement for seniors.", 25.00, "CUSTOMER_LOCATION", []string{"Senior conversation", "Reading", "Board games", "Accompanied walks", "Social stimulation"}},
			{"Respite Care", "Temporary relief and coverage for primary family caregivers for evenings, weekends, or scheduled breaks.", 35.00, "CUSTOMER_LOCATION", []string{"Family caregiver relief", "Overnight respite", "Weekend coverage"}},
			{"Disability Support", "Dedicated personal support for individuals with physical or cognitive disabilities to foster independence.", 35.00, "CUSTOMER_LOCATION", []string{"Adaptive mobility aid", "Independent living support", "Transfer assistance"}},
			{"Care Coordination", "Navigating medical appointments, coordinating physical therapy, care plan setup, and family updates.", 50.00, "CUSTOMER_LOCATION", []string{"Medical appointment accompaniment", "Care plan management", "Family advocacy"}},
		},
	},
	{
		Name:        "Pet Services",
		Description: "Mobile dog grooming, pet sitting, daily dog walking, pet boarding, behavioral training, and vet transport.",
		Image:       "categories/pet_services.jpg",
		Services: []ServiceData{
			{"Pet Grooming", "Mobile van dog and cat grooming, hydrobaths, breed haircuts, deshedding, and nail clipping.", 60.00, "CUSTOMER_LOCATION", []string{"Mobile dog grooming", "Breed haircut", "Deshedding treatment", "Nail trim", "Cat grooming", "Flea bath"}},
			{"Pet Sitting", "In-home pet visits, cat feeding and playtime, puppy check-ins, and overnight stay pet sitting.", 25.00, "CUSTOMER_LOCATION", []string{"In-home pet sitting", "Cat drop-in visits", "Puppy care", "Overnight sitting"}},
			{"Dog Walking", "Private dog walks, energetic group pack walks, neighborhood trail excursions, and bathroom breaks.", 20.00, "CUSTOMER_LOCATION", []string{"Private dog walks", "Group pack walks", "Trail adventures", "Exercise relief walks"}},
			{"Pet Boarding and Daycare", "Cage-free dog boarding, doggy daycare, supervised group play, and cozy overnight accommodations.", 40.00, "PROVIDER_LOCATION", []string{"Cage-free dog boarding", "Doggy daycare", "Supervised playgroup"}},
			{"Pet Training", "Puppy obedience, leash manners, aggressive behavior modification, and trick agility training.", 70.00, "CUSTOMER_LOCATION", []string{"Puppy obedience", "Leash reactivity", "Agility training", "Behavioral correction"}},
			{"Veterinary Care", "Mobile vet visits, wellness checkups, core vaccinations, minor sickness exams, and health certificates.", 85.00, "CUSTOMER_LOCATION", []string{"Mobile vet visits", "Routine pet checkups", "Vaccinations", "Wellness exams"}},
			{"Pet Transportation", "Safe pet taxi services for vet appointments, airport transport, and groomer drop-off and pickup.", 30.00, "CUSTOMER_LOCATION", []string{"Vet trip rides", "Airport pet transport", "Groomer drop-off and pickup"}},
		},
	},
	{
		Name:        "Education and Lessons",
		Description: "Academic tutoring, SAT/ACT test prep, language learning, music lessons, art classes, and coding lessons.",
		Image:       "categories/education.jpg",
		Services: []ServiceData{
			{"Academic Tutoring", "One-on-one K-12 and college tutoring in math, science, reading, writing, history, and chemistry.", 45.00, "ONLINE", []string{"Reading", "Mathematics", "Science", "Specific grades", "Chemistry", "Physics", "History"}},
			{"Test Preparation", "Strategic coaching and practice exams for SAT, ACT, GRE, GMAT, AP tests, and state exams.", 60.00, "ONLINE", []string{"SAT prep", "ACT prep", "GRE/GMAT", "AP exam prep", "State standardized tests"}},
			{"Language Lessons", "Conversational and fluent instruction in Spanish, French, Mandarin, English ESL, Arabic, and German.", 40.00, "ONLINE", []string{"Spanish", "French", "Mandarin", "English ESL", "Arabic", "German", "Sign language"}},
			{"Music Lessons", "Private instrument instruction for piano, acoustic/electric guitar, violin, drums, and vocal training.", 45.00, "CUSTOMER_LOCATION", []string{"Piano lessons", "Guitar lessons", "Vocal coaching", "Violin", "Drum instruction"}},
			{"Art Lessons", "Instruction in drawing, oil painting, watercolor, ceramics, sculpting, and digital graphic illustration.", 40.00, "CUSTOMER_LOCATION", []string{"Drawing", "Oil painting", "Watercolor", "Ceramics", "Digital illustration"}},
			{"Dance Lessons", "Private and group dance coaching in ballet, hip hop, salsa, ballroom, tap, and contemporary dance.", 45.00, "PROVIDER_LOCATION", []string{"Ballet", "Hip hop", "Salsa and bachata", "Contemporary dance", "Ballroom"}},
			{"Computer and Digital Skills", "Coding lessons for kids and adults (Python, JS), Excel mastery, senior digital literacy, and computer basics.", 50.00, "ONLINE", []string{"Coding for kids", "Senior computer basics", "MS Office / Excel", "Graphic software"}},
			{"Vocational and Trade Training", "Foundational preparation and mentoring for electrical, carpentry, HVAC, plumbing, and culinary trades.", 60.00, "PROVIDER_LOCATION", []string{"Electrical apprentice prep", "Carpentry basics", "Culinary skills", "Auto mechanics"}},
		},
	},
	{
		Name:        "Healthcare and Mental Health",
		Description: "Primary healthcare, dentistry, optometry, in-home nursing, physical therapy, counseling, and medical testing.",
		Image:       "categories/healthcare.jpg",
		Services: []ServiceData{
			{"Primary Care", "Routine medical wellness exams, chronic disease management, vitals monitoring, and health consults.", 100.00, "CUSTOMER_LOCATION", []string{"Routine health checkup", "Blood pressure monitoring", "Wellness visits", "Preventative care"}},
			{"Dental Care", "Dental cleanings, oral examinations, teeth whitening, fillings, and mobile dental van care.", 90.00, "PROVIDER_LOCATION", []string{"Dental exams", "Teeth cleaning", "Fluoride treatment", "Cavity fillings", "Teeth whitening"}},
			{"Eye Care", "Comprehensive vision testing, eyeglasses fittings, contact lens exams, and glaucoma screenings.", 75.00, "PROVIDER_LOCATION", []string{"Vision exams", "Prescription glasses", "Contact lens fitting", "Glaucoma screening"}},
			{"Home Nursing", "Skilled in-home nursing, post-op wound care, IV infusions, injections, and medication management.", 65.00, "CUSTOMER_LOCATION", []string{"Wound care", "Post-operative care", "IV infusion", "Medication administration"}},
			{"Physical Therapy", "Musculoskeletal rehabilitation, joint mobility, sports recovery, and post-surgery physical therapy.", 85.00, "CUSTOMER_LOCATION", []string{"Post-surgery rehab", "Sports injury recovery", "Joint mobility", "Chronic back pain"}},
			{"Occupational Therapy", "Rehabilitation to restore daily living abilities, fine motor skills, and stroke recovery support.", 85.00, "CUSTOMER_LOCATION", []string{"Hand therapy", "Activities of daily living rehab", "Stroke recovery"}},
			{"Speech Therapy", "Diagnosis and therapy for speech delays, articulation, fluency, swallowing disorders, and stuttering.", 80.00, "CUSTOMER_LOCATION", []string{"Articulation therapy", "Language development", "Swallowing therapy", "Stuttering"}},
			{"Counseling and Psychotherapy", "Licensed individual, couples, and family therapy for anxiety, depression, grief, and life transitions.", 90.00, "ONLINE", []string{"Cognitive behavioral therapy", "Anxiety and depression", "Couples counseling", "Family therapy"}},
			{"Psychiatry", "Psychiatric medical evaluations, psychiatric consultations, and safe prescription medication management.", 140.00, "ONLINE", []string{"Psychiatric evaluations", "Medication management", "Mental wellness consultations"}},
			{"Nutrition and Dietitian Services", "Personalized nutritional plans, medical diet consulting, weight management, and meal guidance.", 70.00, "ONLINE", []string{"Weight management", "Diabetic meal plans", "Sports nutrition", "Gut health diets"}},
			{"Medical Testing and Imaging", "In-home blood draws (phlebotomy), mobile X-rays, ultrasound imaging, and diagnostic test panels.", 95.00, "CUSTOMER_LOCATION", []string{"Mobile blood draws", "X-ray imaging", "Ultrasound scans", "Rapid diagnostic tests"}},
			{"Specialty Medical Care", "Specialist consultations across dermatology, podiatry, cardiology, and allergy care.", 125.00, "PROVIDER_LOCATION", []string{"Dermatology consultations", "Podiatry foot care", "Cardiology checkups", "Allergy care"}},
		},
	},
	{
		Name:        "Food and Catering",
		Description: "Catering for events, in-home personal chefs, custom bakeries, meal prep delivery, and dessert bars.",
		Image:       "categories/food_catering.jpg",
		Services: []ServiceData{
			{"Catering", "Full-service wedding and corporate catering, hot buffet stations, box lunches, and BBQ catering.", 30.00, "CUSTOMER_LOCATION", []string{"Corporate buffet catering", "Wedding catering", "Box lunches", "Barbecue catering", "Plated dinners"}},
			{"Personal Chef Services", "Private in-home chef for intimate dinner parties, romantic dates, custom menus, and culinary events.", 75.00, "CUSTOMER_LOCATION", []string{"In-home private chef", "Multi-course dinner parties", "Romantic dinners", "Dietary custom cooking"}},
			{"Meal Preparation", "Weekly customized healthy meal prep, portion-controlled containers, keto, vegan, and family meals.", 50.00, "CUSTOMER_LOCATION", []string{"Weekly meal prep", "Keto and low carb prep", "Vegan meal prep", "Family dinners"}},
			{"Bakery Services", "Custom birthday cakes, wedding cakes, fresh artisan breads, donuts, pastries, and gluten-free baking.", 40.00, "PROVIDER_LOCATION", []string{"Donuts", "Birthday cakes", "Wedding cakes", "Pastries", "Artisan bread", "Gluten-free baking"}},
			{"Desserts and Frozen Treats", "Ice cream carts, cupcake towers, gourmet cookies, dessert tables, churro stations, and crepe bars.", 35.00, "CUSTOMER_LOCATION", []string{"Ice cream carts", "Custom cookies", "Cupcakes", "Dessert tables", "Churros and crepes"}},
			{"Coffee and Beverage Services", "Mobile espresso bars, barista service, craft mocktail stations, fresh smoothie carts, and boba tea.", 45.00, "CUSTOMER_LOCATION", []string{"Mobile espresso bar", "Smoothies and juice bar", "Boba tea bar", "Craft mocktails"}},
			{"Restaurants and Prepared Meals", "Group family feast packages, artisan takeout catering platters, and hot prepared meals.", 25.00, "PROVIDER_LOCATION", []string{"Takeout catering", "Family feast platters", "Gourmet ready-to-eat dishes"}},
		},
	},
	{
		Name:        "Events and Entertainment",
		Description: "Event coordination, party rentals, floral decor, live DJs, musicians, performers, and professional bartending.",
		Image:       "categories/events_entertainment.jpg",
		Services: []ServiceData{
			{"Event Planning", "Full-service wedding planning, corporate gala coordination, birthday party management, and design.", 80.00, "CUSTOMER_LOCATION", []string{"Wedding planning", "Corporate galas", "Birthday party coordination", "Full-service event design"}},
			{"Event Decoration", "Balloon arches, floral centerpieces, photo backdrops, mood lighting, and aesthetic table styling.", 60.00, "CUSTOMER_LOCATION", []string{"Balloon arches", "Floral arrangements", "Backdrops and photo booths", "Table centerpieces"}},
			{"Event Rentals", "Tables, chairs, party tents, dance floors, linens, staging, and audiovisual rental equipment.", 50.00, "CUSTOMER_LOCATION", []string{"Tables and chairs", "Tent rentals", "Linen and tableware", "Dance floors", "Audio/Visual gear"}},
			{"Venue Rental", "Private event venues, scenic gardens, banquet halls, rustic barns, and modern urban loft spaces.", 150.00, "PROVIDER_LOCATION", []string{"Banquet halls", "Outdoor garden venues", "Loft event spaces", "Meeting rooms"}},
			{"DJs", "Live event DJs, wedding sound systems, MC hosting, dance party mixes, and intelligent lighting.", 90.00, "CUSTOMER_LOCATION", []string{"Wedding DJ", "Corporate event DJ", "Club and party DJ", "Karaoke hosting"}},
			{"Live Performers", "Live bands, acoustic soloists, string quartets, magicians, comedians, and specialty musicians.", 100.00, "CUSTOMER_LOCATION", []string{"Live bands", "Acoustic soloists", "String quartets", "Magicians", "Vocalists"}},
			{"Party Entertainment", "Face painters, balloon artists, costumed character visits, clowns, and inflatable bounce houses.", 55.00, "CUSTOMER_LOCATION", []string{"Face painting", "Balloon twisting", "Character visits", "Clowns", "Inflatable bounce houses"}},
			{"Bartending", "Certified event mixologists, mobile pop-up bar setups, signature cocktail creation, and beverage pouring.", 45.00, "CUSTOMER_LOCATION", []string{"Mobile bar service", "Craft cocktail mixologists", "Beer and wine pouring", "Event bar setup"}},
		},
	},
	{
		Name:        "Photography and Creative Services",
		Description: "Professional photography, videography, photo retouching, graphic design, copywriting, and printing.",
		Image:       "categories/photography.jpg",
		Services: []ServiceData{
			{"Photography", "Wedding photography, family portraits, newborn sessions, corporate headshots, and real estate photos.", 95.00, "CUSTOMER_LOCATION", []string{"Weddings", "Family portraits", "Newborns", "Events", "Product photography", "Headshots", "Real estate"}},
			{"Videography", "Cinematic wedding films, corporate promos, drone aerial videography, social media reels, and event videos.", 120.00, "CUSTOMER_LOCATION", []string{"Wedding films", "Commercial video", "Drone aerial footage", "Social media reels", "Event recaps"}},
			{"Photo and Video Editing", "Professional photo retouching, color grading, sound design, video cutting, and album creation.", 50.00, "ONLINE", []string{"Color grading", "Retouching", "Video post-production", "Album design", "Sound editing"}},
			{"Graphic Design and Branding", "Logo design, comprehensive brand guidelines, business cards, flyers, and digital promotional graphics.", 55.00, "ONLINE", []string{"Logo design", "Brand identity", "Business cards", "Brochures and flyers", "UI design"}},
			{"Writing and Editing", "Professional copywriting, website content, resume writing, blog posts, proofreading, and ghostwriting.", 45.00, "ONLINE", []string{"Copywriting", "Technical writing", "Proofreading", "Blog articles", "Ghostwriting"}},
			{"Printing and Publishing", "Large format poster printing, custom vinyl banners, branded apparel, business brochures, and book binding.", 40.00, "PROVIDER_LOCATION", []string{"Banner printing", "Custom apparel printing", "Large format prints", "Book binding"}},
		},
	},
	{
		Name:        "Business and Professional Services",
		Description: "Accounting, tax preparation, legal contracts, notary public, consulting, real estate, and digital marketing.",
		Image:       "categories/business_services.jpg",
		Services: []ServiceData{
			{"Accounting and Bookkeeping", "Monthly business bookkeeping, QuickBooks management, financial reporting, and payroll processing.", 65.00, "ONLINE", []string{"Quickbooks setup", "Monthly bookkeeping", "Financial statement preparation", "Payroll"}},
			{"Tax Preparation", "Individual tax return filing, corporate business taxes, tax deduction planning, and IRS advisory.", 85.00, "ONLINE", []string{"Individual tax returns", "Business corporate taxes", "Tax planning", "IRS resolution"}},
			{"Legal Services", "Business contract drafting, LLC formation, estate planning, wills, and trademark registration.", 150.00, "ONLINE", []string{"Contract review", "Business formation", "Estate planning and wills", "Trademark filing"}},
			{"Notary Services", "Mobile notary public, loan signing agent, document witnessing, apostille services, and power of attorney.", 45.00, "CUSTOMER_LOCATION", []string{"Mobile notary public", "Loan signing agent", "Document witnessing", "Apostille assistance"}},
			{"Business Consulting", "Strategic business growth, startup launch advisory, operations efficiency, and financial modeling.", 100.00, "ONLINE", []string{"Strategic planning", "Startup advisory", "Operations optimization", "Franchise consulting"}},
			{"Real Estate Services", "Home buying/selling representation, property management, home staging, and rental leasing.", 100.00, "CUSTOMER_LOCATION", []string{"Buyer representation", "Home staging", "Property management", "Rental leasing"}},
			{"Resume and Career Services", "Executive resume writing, LinkedIn profile optimization, interview preparation, and career coaching.", 60.00, "ONLINE", []string{"Resume writing", "LinkedIn profile optimization", "Interview coaching", "Career coaching"}},
			{"Translation and Interpretation", "Certified document translation, legal transcription, and real-time live language interpretation.", 50.00, "ONLINE", []string{"Document translation", "In-person court interpretation", "Medical translation"}},
			{"Virtual Assistance", "Remote administrative assistance, calendar scheduling, inbox management, and data entry.", 25.00, "ONLINE", []string{"Email and calendar management", "Data entry", "Customer service support", "Research tasks"}},
			{"Marketing Services", "Social media management, SEO optimization, Google and Facebook paid ads, and email campaigns.", 75.00, "ONLINE", []string{"Social media", "SEO", "Paid advertising", "Email marketing", "Influencer outreach", "Content creation"}},
		},
	},
	{
		Name:        "Technology Services",
		Description: "Device repair, IT technical support, home Wi-Fi networks, TV mounting, smart home setup, and web/app development.",
		Image:       "categories/technology.jpg",
		Services: []ServiceData{
			{"Computer and Device Repair", "PC and Mac repair, laptop screen replacement, virus removal, hardware upgrades, and data recovery.", 70.00, "CUSTOMER_LOCATION", []string{"Laptop screen repair", "PC troubleshooting", "Virus and malware removal", "Data recovery"}},
			{"Technical Support", "Remote and on-site IT help, software installations, printer troubleshooting, and operating system updates.", 55.00, "ONLINE", []string{"Printer setup", "Remote desktop support", "Software installation", "OS upgrades"}},
			{"Home Networking and Wi Fi", "Mesh Wi-Fi system setup, router optimization, Ethernet cabling, and dead-zone elimination.", 75.00, "CUSTOMER_LOCATION", []string{"Mesh Wi-Fi setup", "Ethernet cabling", "Router configuration", "Dead zone elimination"}},
			{"TV and Home Theater Installation", "Wall TV mounting, concealed wiring, soundbar installation, surround sound, and home cinema setup.", 85.00, "CUSTOMER_LOCATION", []string{"TV wall mounting", "Surround sound setup", "Projector installation", "Wire concealment"}},
			{"Smart Home and Security Installation", "Video doorbell installation, security camera setup, smart thermostats, smart locks, and home automation.", 90.00, "CUSTOMER_LOCATION", []string{"Ring doorbell setup", "Smart thermostat", "Security camera installation", "Smart locks"}},
			{"Website Development", "Custom responsive website design, WordPress sites, Shopify e-commerce, and web application development.", 80.00, "ONLINE", []string{"WordPress websites", "E-commerce Shopify stores", "Custom web applications", "Responsive design"}},
			{"App and Software Development", "iOS and Android mobile app development, backend APIs, database architecture, and custom software systems.", 100.00, "ONLINE", []string{"iOS and Android mobile apps", "Backend API development", "Database design", "Cloud deployment"}},
			{"Automation and Systems Integration", "Zapier workflow automation, CRM integrations, API webhook connections, and AI chatbot setups.", 90.00, "ONLINE", []string{"Zapier and webhook automation", "CRM integration", "API workflows", "AI chatbot integration"}},
		},
	},
}

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
