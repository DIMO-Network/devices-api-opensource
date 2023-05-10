package services

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	ethC "github.com/ethereum/go-ethereum/crypto"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared/db"
	"github.com/twinj/uuid"
)

type VerifiableCredentialService struct {
	dbs        db.Store
	ID         string
	Issuer     string
	PublicKey  ecdsa.PublicKey
	PrivateKey ecdsa.PrivateKey
}

// NewVerifiableCredentialService generates vc service
func NewVerifiableCredentialService(dbs db.Store, settings *config.Settings) (*VerifiableCredentialService, error) {
	xb, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyX)
	if err != nil {
		return nil, err
	}
	x := new(big.Int).SetBytes(xb)

	yb, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyY)
	if err != nil {
		return nil, err
	}
	y := new(big.Int).SetBytes(yb)

	db, err := base64.RawURLEncoding.DecodeString(settings.DIMOKeyD)
	if err != nil {
		return nil, err
	}
	d := new(big.Int).SetBytes(db)

	publicKey := ecdsa.PublicKey{
		Curve: ethC.S256(),
		X:     x,
		Y:     y,
	}
	privateKey := ecdsa.PrivateKey{
		PublicKey: publicKey,
		D:         d,
	}

	return &VerifiableCredentialService{
		dbs:        dbs,
		ID:         "did:dimo_example:abfe13f712120431c276e12ecab",
		Issuer:     "dimo.zone",
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil
}

type Claim struct {
	ID      string
	Vin     string
	TokenID string
}

type Proof struct {
	TypeOfProof string
	Created     time.Time
	Creator     crypto.PublicKey
	Signature   struct {
		R, S *big.Int
	}
}

type Credential struct {
	Context           []string
	ID                string
	TypeOfCredential  []string
	Issuer            string
	IssuanceDate      time.Time
	CredentialSubject Claim
	Proof             Proof
	Status            Status
}

type Status struct {
	ID     string
	Status string
}

var secp256k1Prefix = []byte{0xe7, 0x01}

//create credential
func (vcs *VerifiableCredentialService) createVinCredential(vin, tokenID string) (*Credential, error) {
	claim := Claim{
		ID:      "did:uuid:" + uuid.NewV1().String(),
		Vin:     vin,
		TokenID: tokenID,
	}
	credential := Credential{
		Context:           []string{"https://raw.githubusercontent.com/DIMO-Network/dimo-vin-kyc/main/dimo-vin-kyc.json", "https://raw.githubusercontent.com/DIMO-Network/dimo-vin-kyc/main/dimo-vin-kyc.jsonld"},
		ID:                vcs.ID,
		TypeOfCredential:  []string{"VerifiableCredential", "VinVerification"},
		Issuer:            vcs.Issuer,
		IssuanceDate:      time.Now(),
		CredentialSubject: claim,
	}
	r, s, err := ecdsa.Sign(rand.Reader, &vcs.PrivateKey, []byte(fmt.Sprintf("%v", credential)))
	if err != nil {
		return nil, err
	}
	//create proof
	proof := Proof{
		TypeOfProof: "ed25519",
		Created:     time.Now(),
		Creator:     vcs.PublicKey,
		Signature: struct {
			R *big.Int
			S *big.Int
		}{R: r, S: s},
	}
	//add proof
	credential.Proof = proof
	return &credential, nil
}

func (vcs *VerifiableCredentialService) VerifyCredential(tokenID string, credential Credential) (bool, error) {
	return ecdsa.Verify(&vcs.PublicKey, []byte(fmt.Sprintf("%v", credential)), credential.Proof.Signature.R, credential.Proof.Signature.S), nil
}

// //create presentation
// func createPresentation(keyPair KeyPair, metadata PresenterMetadata, credential Credential) Presentation {
// 	presentation := Presentation{
// 		context:           metadata.context,
// 		typeOfPresentaion: metadata.typeOfPresentation,
// 		credential:        credential,
// 	}

// 	//create proof
// 	proofOfPresentaton := Proof{
// 		typeOfProof: "ed25519",
// 		created:     time.Now(),
// 		creator:     keyPair.publicKey,
// 		signature:   ed25519.Sign(keyPair.privateKey, []byte(fmt.Sprintf("%v", presentation))),
// 	}

// 	presentation.proof = proofOfPresentaton

// 	return presentation

// }

// //verify presentation
// func verifyPresentation(publicKey ed25519.PublicKey, presentation Presentation) bool {
// 	//verify if the public key is the same as in the credential
// 	if string(publicKey) != string(presentation.proof.creator) {
// 		return false
// 	}
// 	proofObj := presentation.proof
// 	presentation.proof = Proof{}
// 	//verify signature
// 	return ed25519.Verify(publicKey, []byte(fmt.Sprintf("%v", presentation)), proofObj.signature)
// }
