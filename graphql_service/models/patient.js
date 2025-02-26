var mongoose = require('mongoose');
var Schema = mongoose.Schema;

var PatientSchema = new Schema({
    _id: {
        type: String,
        required: true
    },
    resourceType: {
        type: String,
        required: true
    },
    name: {
        use: {
            type:String,
            required:true
        },
        family: {
            type:String,
            required:true
        },
        given: [String],
        prefix: [String]
    },
    birthDate: {
        time: Date
    },
    generalPractitioner: {
        reference: {
            type: String,
            required: true
        }
    }

});
var PatientModel = mongoose.model('patients', PatientSchema);
// module.exports = PatientModel;

exports.find_by_id = (patient_id) => {
    var patient = PatientModel.findOne({"_id": patient_id});
    return patient
}

exports.find_by_ids = (patient_id_list, limit) => {
    var patient = PatientModel.find({"_id" : { $in: patient_id_list }}).limit(limit)
    return patient
}

exports.find_by_name = (patient_name, limit) => {
    var patients = PatientModel.find({ "name": { $elemMatch: { "given": args.name } } }).limit(limit);
    return patients
}

exports.find = (limit) => {
    var patients = PatientModel.find().limit(limit);
    return patients
}