$(function() {

  var activeModal;

  var modalToggle = $('[data-toggle="modal"]'),
      modalClose = $('[data-dismiss="modal"]');

  modalToggle.on('click', function() {
    var target = $(this).data('target');
    activeModal = $(target);
    activeModal.show();
  });

  modalClose.on('click', function() {
    activeModal.hide();
    activeModal = null;
  });

});
